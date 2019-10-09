package es

import (
	"context"
	"fmt"
	"testing"

	elastic6 "github.com/elastic/go-elasticsearch/v6"
	elastic7 "github.com/elastic/go-elasticsearch/v7"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchIndexLifecyclePolicy(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIndexLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchIndexLifecyclePolicy,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexLifecyclePolicyExists("elasticsearch_index_lifecycle_policy.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_index_lifecycle_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchIndexLifecyclePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No index lifecycle policy ID is set")
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			res, err := client.API.ILM.GetLifecycle(
				client.API.ILM.GetLifecycle.WithContext(context.Background()),
				client.API.ILM.GetLifecycle.WithPretty(),
				client.API.ILM.GetLifecycle.WithPolicy(rs.Primary.ID),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				return errors.Errorf("Error when get lifecycle policy %s: %s", rs.Primary.ID, res.String())
			}
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			res, err := client.API.ILM.GetLifecycle(
				client.API.ILM.GetLifecycle.WithContext(context.Background()),
				client.API.ILM.GetLifecycle.WithPretty(),
				client.API.ILM.GetLifecycle.WithPolicy(rs.Primary.ID),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				return errors.Errorf("Error when get lifecycle policy %s: %s", rs.Primary.ID, res.String())
			}
		default:
			return errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
		}

		return nil
	}
}

func testCheckElasticsearchIndexLifecyclePolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index_lifecycle_policy" {
			continue
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			res, err := client.API.ILM.GetLifecycle(
				client.API.ILM.GetLifecycle.WithContext(context.Background()),
				client.API.ILM.GetLifecycle.WithPretty(),
				client.API.ILM.GetLifecycle.WithPolicy(rs.Primary.ID),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				if res.StatusCode == 404 {
					return nil
				}
			}
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			res, err := client.API.ILM.GetLifecycle(
				client.API.ILM.GetLifecycle.WithContext(context.Background()),
				client.API.ILM.GetLifecycle.WithPretty(),
				client.API.ILM.GetLifecycle.WithPolicy(rs.Primary.ID),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				if res.StatusCode == 404 {
					return nil
				}
			}
		default:
			return errors.New("Index Lifecycle Management is only supported by the elastic library >= v6!")
		}

		return fmt.Errorf("Index lifecycle policy %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchIndexLifecyclePolicy = `
resource "elasticsearch_index_lifecycle_policy" "test" {
  name = "terraform-test"
  policy = <<EOF
{
  "policy": {
    "phases": {
      "warm": {
        "min_age": "10d",
        "actions": {
          "forcemerge": {
            "max_num_segments": 1
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
EOF
}
`
