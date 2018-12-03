package igw

import (
	"waze/terraform/aws_terraforming/aws_generator"
	"waze/terraform/terraform_utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var ignoreKey = map[string]bool{
	"^id$": true,
}

var allowEmptyValues = map[string]bool{
	"tags.": true,
}

type IgwGenerator struct {
	aws_generator.BasicGenerator
}

// Generate TerraformResources from AWS API,
// from each Internet gateway create 1 TerraformResource.
// Need InternetGatewayId as ID for terraform resource
func (IgwGenerator) createResources(igws *ec2.DescribeInternetGatewaysOutput) []terraform_utils.TerraformResource {
	resoures := []terraform_utils.TerraformResource{}
	for _, internetGateway := range igws.InternetGateways {
		resourceName := ""
		if len(internetGateway.Tags) > 0 {
			for _, tag := range internetGateway.Tags {
				if aws.StringValue(tag.Key) == "Name" {
					resourceName = aws.StringValue(tag.Value)
					break
				}
			}
		}
		resoures = append(resoures, terraform_utils.TerraformResource{
			ResourceType: "aws_internet_gateway",
			ResourceName: resourceName,
			Item:         nil,
			ID:           aws.StringValue(internetGateway.InternetGatewayId),
			Provider:     "aws",
		})
	}
	return resoures
}

func (g IgwGenerator) Generate(region string) ([]terraform_utils.TerraformResource, map[string]terraform_utils.ResourceMetaData, error) {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String(region)})
	svc := ec2.New(sess)
	igws, err := svc.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{})
	if err != nil {
		return []terraform_utils.TerraformResource{}, map[string]terraform_utils.ResourceMetaData{}, err
	}
	resources := g.createResources(igws)
	metadata := terraform_utils.NewResourcesMetaData(resources, ignoreKey, allowEmptyValues, map[string]string{})
	return resources, metadata, nil

}
