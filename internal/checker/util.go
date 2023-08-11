package checker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"unicode"

	"github.com/bufbuild/protocompile"
	"github.com/oshokin/protolinter/internal/common"
	"github.com/oshokin/protolinter/internal/config"
	"github.com/oshokin/protolinter/internal/logger"
)

const (
	githubLinkPartsCount      = 4
	googleProtobufPrefix      = "google/protobuf"
	googleAPIPrefix           = "google/api/"
	protocGenOpenAPIV2Prefix  = "protoc-gen-openapiv2/"
	googleAPIsGitHubPath      = "github.com/googleapis/googleapis"
	grpcGatewayGitHubPath     = "github.com/grpc-ecosystem/grpc-gateway"
	githubDomain              = "github.com/"
	githubDownloadLinkPattern = "https://raw.githubusercontent.com/%s/%s/master/%s"
)

func getSourceResolver(ctx context.Context, cfg *config.Config) *protocompile.SourceResolver {
	return &protocompile.SourceResolver{
		Accessor: func(path string) (io.ReadCloser, error) {
			_, err := os.Stat(path)
			if err == nil || strings.HasPrefix(path, googleProtobufPrefix) {
				return os.Open(path)
			}

			switch {
			case strings.HasPrefix(path, googleAPIPrefix):
				path, err = url.JoinPath(googleAPIsGitHubPath, path)
				if err != nil {
					return nil, err
				}
			case strings.HasPrefix(path, protocGenOpenAPIV2Prefix):
				path, err = url.JoinPath(grpcGatewayGitHubPath, path)
				if err != nil {
					return nil, err
				}
			}

			resource := getDownloadLink(path)
			if cfg.GetVerboseMode() {
				logger.Warnf(ctx, "Downloading proto dependency, %s: %s, %s: %s",
					common.FileNameTag, path,
					common.URLTag, resource)
			}

			request, err := http.NewRequestWithContext(ctx, http.MethodGet, resource, nil)
			if err != nil {
				return nil, err
			}

			response, err := http.DefaultClient.Do(request)
			if err != nil {
				return nil, err
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				return nil, err
			}

			return io.NopCloser(bytes.NewReader(body)), nil
		},
	}
}

func getDownloadLink(importPath string) string {
	if !strings.HasPrefix(importPath, githubDomain) {
		return importPath
	}

	parts := strings.SplitN(importPath, "/", githubLinkPartsCount)
	if len(parts) < githubLinkPartsCount {
		return importPath
	}

	var (
		user     = parts[1]
		repo     = parts[2]
		filePath = parts[3]
	)

	return fmt.Sprintf(githubDownloadLinkPattern, user, repo, filePath)
}

func startsWithCapitalLetter(s string) bool {
	if len(s) == 0 {
		return false
	}

	return unicode.IsUpper(rune(s[0]))
}
