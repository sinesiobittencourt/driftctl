package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsRoute(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Paths: []string{"./testdata/acc/aws_route"},
		Args:  []string{"scan", "--filter", "Type=='aws_route'"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.Equal(6, result.Summary().TotalManaged)
				},
			},
		},
	})
}
