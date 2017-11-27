package reqsigssm

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/remind101/pkg/httpx/middleware"
)

// NewWithPath takes a path, and returns an instance of
// middleware.StaticSigningKeyRepository initialized with client keys / secrets
// fetched from the given path. The assumption is that path points to an SSM
// directory containing the client keys and secrets.
func NewWithPath(path string) (*middleware.StaticSigningKeyRepository, error) {
	kmap := map[string]string{}
	sess := session.Must(session.NewSession())
	svc := ssm.New(sess)
	in := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: aws.Bool(true),
	}

	err := svc.GetParametersByPathPages(in, func(out *ssm.GetParametersByPathOutput, cont bool) bool {
		for _, param := range out.Parameters {
			kmap[aws.StringValue(param.Name)] = aws.StringValue(param.Value)
		}
		return true
	})
	r := middleware.NewStaticSigningKeyRepository(kmap)

	return r, err
}
