package github

import (
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourcegithub "github.com/cloudskiff/driftctl/pkg/resource/github"
	ghdeserializer "github.com/cloudskiff/driftctl/pkg/resource/github/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type GithubRepositorySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repository   GithubRepository
	runner       *terraform.ParallelResourceReader
}

func NewGithubRepositorySupplier(provider *GithubTerraformProvider, repository GithubRepository) *GithubRepositorySupplier {
	return &GithubRepositorySupplier{
		provider,
		ghdeserializer.NewGithubRepositoryDeserializer(),
		repository,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s GithubRepositorySupplier) Resources() ([]resource.Resource, error) {

	resourceList, err := s.repository.ListRepositories()

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourcegithub.GithubRepositoryResourceType)
	}

	for _, id := range resourceList {
		id := id
		s.runner.Run(func() (cty.Value, error) {
			completeResource, err := s.reader.ReadResource(terraform.ReadResourceArgs{
				Ty: resourcegithub.GithubRepositoryResourceType,
				ID: id,
			})
			if err != nil {
				logrus.Warnf("Error reading %s[%s]: %+v", id, resourcegithub.GithubRepositoryResourceType, err)
				return cty.NilVal, err
			}
			return *completeResource, nil
		})
	}

	results, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(results)
}
