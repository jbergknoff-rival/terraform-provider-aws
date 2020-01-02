package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEcsDataSource_ecsTaskDefinition(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsTaskDefinitionDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ecs_task_definition.mongo", "family", rName),
					resource.TestCheckResourceAttr("data.aws_ecs_task_definition.mongo", "network_mode", "bridge"),
					resource.TestMatchResourceAttr("data.aws_ecs_task_definition.mongo", "revision", regexp.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr("data.aws_ecs_task_definition.mongo", "status", "ACTIVE"),
					resource.TestMatchResourceAttr("data.aws_ecs_task_definition.mongo", "task_role_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:iam::[^:]+:role/%s$", rName))),
				),
			},
		},
	})
}

func TestAccAWSEcsDataSource_ecsTaskDefinition_nonexistent(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-missing-%s", acctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsNonexistentTaskDefinitionDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ecs_task_definition.nonexistent", "family", rName),
					resource.TestCheckResourceAttr("data.aws_ecs_task_definition.nonexistent", "revision", "0"),
					resource.TestCheckNoResourceAttr("data.aws_ecs_task_definition.nonexistent", "task_role_arn"),
				),
			},
		},
	})
}

func testAccCheckAwsEcsTaskDefinitionDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "mongo_role" {
  name = "%[1]s"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_ecs_task_definition" "mongo" {
  family        = "%[1]s"
  task_role_arn = "${aws_iam_role.mongo_role.arn}"
  network_mode  = "bridge"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "environment": [{
      "name": "SECRET",
      "value": "KEY"
    }],
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "memoryReservation": 64,
    "name": "mongodb"
  }
]
DEFINITION
}

data "aws_ecs_task_definition" "mongo" {
  task_definition = "${aws_ecs_task_definition.mongo.family}"
}
`, rName)
}

func testAccCheckAwsEcsNonexistentTaskDefinitionDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ecs_task_definition" "nonexistent" {
  task_definition = "%[1]s"
  missing_okay = true
}
`, rName)
}
