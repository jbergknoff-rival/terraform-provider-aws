package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsEcsTaskDefinition() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsEcsTaskDefinitionRead,

		Schema: map[string]*schema.Schema{
			"task_definition": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"missing_okay": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			// Computed values.
			"family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"task_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEcsTaskDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ecsconn

	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(d.Get("task_definition").(string)),
	}
	log.Printf("[DEBUG] Reading ECS Task Definition: %s", params)
	desc, err := conn.DescribeTaskDefinition(params)

	if err != nil {
		if d.Get("missing_okay").(bool) && strings.Contains(err.Error(), "Unable to describe task definition") {
			d.SetId("missing-task-definition-" + d.Get("task_definition").(string))
			d.Set("family", d.Get("task_definition").(string))
			// The common technique for having Terraform manage config (e.g. environment variables), while
			// deploying outside of Terraform, involves setting a service's task definition reference to
			//
			// max(data.aws_ecs_task_definition.x.revision, aws_ecs_task_definition.x.revision)
			//
			// In a bootstrapping situation, where the task definition doesn't yet exist, falling back on a revision
			// of 0 will make this create the resource and use that revision.
			d.Set("revision", 0)
			return nil
		}

		return fmt.Errorf("Failed getting task definition %s %q", err, d.Get("task_definition").(string))
	}

	taskDefinition := *desc.TaskDefinition

	d.SetId(aws.StringValue(taskDefinition.TaskDefinitionArn))
	d.Set("family", aws.StringValue(taskDefinition.Family))
	d.Set("network_mode", aws.StringValue(taskDefinition.NetworkMode))
	d.Set("revision", aws.Int64Value(taskDefinition.Revision))
	d.Set("status", aws.StringValue(taskDefinition.Status))
	d.Set("task_role_arn", aws.StringValue(taskDefinition.TaskRoleArn))

	if d.Id() == "" {
		return fmt.Errorf("task definition %q not found", d.Get("task_definition").(string))
	}

	return nil
}
