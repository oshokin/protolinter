package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/oshokin/protolinter/internal/common"
	"github.com/oshokin/protolinter/internal/config"
	"github.com/oshokin/protolinter/internal/logger"
)

const (
	githubLinkPartsCount      = 4
	googlePrefix              = "google/"
	googleProtobufPrefix      = "google/protobuf"
	protocGenOpenAPIV2Prefix  = "protoc-gen-openapiv2/"
	googleAPIsGitHubPath      = "github.com/googleapis/googleapis"
	grpcGatewayGitHubPath     = "github.com/grpc-ecosystem/grpc-gateway"
	githubDomain              = "github.com"
	githubDomainPrefix        = githubDomain + "/"
	githubDownloadLinkPattern = "https://raw.githubusercontent.com/%s/%s/master/%s"
)

func (c *ProtoChecker) getSourceResolver(ctx context.Context, cfg *config.Config) *protocompile.SourceResolver {
	// Prepare the module name for imports.
	moduleName := cfg.GetModuleName()
	if moduleName != "" {
		moduleName = strings.Join([]string{moduleName, "/"}, "")
	}

	return &protocompile.SourceResolver{
		Accessor: func(path string) (io.ReadCloser, error) {
			// Check if the requested file exists locally.
			_, err := os.Stat(path)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}

			// "google/protobuf" imports are handled by the resolver created with WithStandardImports.
			if err == nil || strings.HasPrefix(path, googleProtobufPrefix) {
				//nolint: gosec // We're only processing inner files.
				return os.Open(path)
			}

			// If the content is already cached, retrieve it from the cache.
			if cachedContent, ok := c.getCachedContent(path); ok {
				return io.NopCloser(bytes.NewReader(cachedContent)), nil
			}

			newPath := path

			switch {
			// Local import within the module.
			case moduleName != "" && strings.HasPrefix(newPath, moduleName):
				newPath = strings.TrimPrefix(newPath, moduleName)
				//nolint: gosec // We're only processing inner files.
				return os.Open(newPath)
			// Imports not covered by the resolver created with WithStandardImports.
			case strings.HasPrefix(newPath, googlePrefix):
				newPath, err = url.JoinPath(googleAPIsGitHubPath, newPath)
				if err != nil {
					return nil, err
				}
			case strings.HasPrefix(newPath, protocGenOpenAPIV2Prefix):
				newPath, err = url.JoinPath(grpcGatewayGitHubPath, newPath)
				if err != nil {
					return nil, err
				}
			}

			resource, isLocalFile := c.getProtoDependencyPath(cfg.GetGitHubURL(), newPath)

			return c.downloadProtoDependency(ctx, path, resource, isLocalFile)
		},
	}
}

func (c *ProtoChecker) getProtoDependencyPath(githubURL, importPath string) (string, bool) {
	// If the custom GitHub URL is provided, replace GitHub links.
	if githubURL != "" && strings.HasPrefix(importPath, githubDomainPrefix) {
		parsedURL, err := url.Parse(githubURL)
		if err != nil {
			resultURL := strings.Replace(importPath, githubDomain, githubURL, 1)

			return resultURL, true
		}

		isSchemeSupported := len(parsedURL.Scheme) > 1

		switch {
		case !isSchemeSupported:
			// TODO: Сделать фикс для Windows и относительных путей
			// https://github.com/go-openapi/spec/blob/master/normalizer_windows.go
			resultURL := strings.Replace(importPath, githubDomain, "", 1)

			return filepath.Join(githubURL, filepath.FromSlash(resultURL)), true
		case parsedURL.Scheme == "file":
			// TODO: Сделать фикс для Windows и относительных путей
			// https://github.com/go-openapi/spec/blob/master/normalizer_windows.go
			githubURL = strings.TrimPrefix(githubURL, "file://")
			resultURL := strings.Replace(importPath, githubDomain, githubURL, 1)

			return resultURL, true
		default:
			return strings.Replace(importPath, githubDomain, githubURL, 1), false
		}
	}

	if !strings.HasPrefix(importPath, githubDomainPrefix) {
		return importPath, false
	}

	parts := strings.SplitN(importPath, "/", githubLinkPartsCount)
	if len(parts) < githubLinkPartsCount {
		return importPath, false
	}

	var (
		user     = parts[1]
		repo     = parts[2]
		filePath = parts[3]
	)

	return fmt.Sprintf(githubDownloadLinkPattern, user, repo, filePath), false
}

func (c *ProtoChecker) downloadProtoDependency(
	ctx context.Context,
	path,
	resource string,
	isLocalFile bool,
) (io.ReadCloser, error) {
	logger.Infof(ctx, "Fetching proto dependency, %s: %s, %s: %s",
		common.FileNameTag, path,
		common.URLTag, resource)

	if isLocalFile {
		//nolint: gosec // We're only processing inner files.
		return os.Open(resource)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, resource, nil)
	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported protocol scheme") {
			err = fmt.Errorf("'%s' is not valid link or local file doesn't exist", resource)
		}

		return nil, err
	}

	defer response.Body.Close()

	return c.processDownloadedProtoDependency(response, path, resource)
}

func (c *ProtoChecker) processDownloadedProtoDependency(
	response *http.Response,
	path,
	resource string,
) (io.ReadCloser, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		errors := c.extractArtifactoryErrors(body)
		if errors == "" {
			errors = string(body)
		}

		if errors != "" {
			return nil, fmt.Errorf(
				"failed to download file from %s for proto dependency %s, status code: %d, errors: %s",
				resource, path, response.StatusCode, errors)
		}

		return nil, fmt.Errorf(
			"failed to download file from %s for proto dependency %s, status code: %d",
			resource, path, response.StatusCode)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf(
			"file downloaded from %s for proto dependency %s is empty",
			resource, path)
	}

	c.storeCachedContent(path, body)

	return io.NopCloser(bytes.NewReader(body)), nil
}

func (c *ProtoChecker) extractArtifactoryErrors(body []byte) string {
	var (
		response ArtifactoryErrorResponse
		err      = json.Unmarshal(body, &response)
	)

	if err != nil || len(response.Errors) == 0 {
		return ""
	}

	var errorList []string
	for _, artifactoryError := range response.Errors {
		errorList = append(errorList,
			fmt.Sprintf("error code: %d, message: %s",
				artifactoryError.Status, artifactoryError.Message))
	}

	return strings.Join(errorList, "\n")
}

func (c *ProtoChecker) storeCachedContent(path string, content []byte) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	c.cache[path] = content
}

func (c *ProtoChecker) getCachedContent(path string) ([]byte, bool) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	content, ok := c.cache[path]

	return content, ok
}
