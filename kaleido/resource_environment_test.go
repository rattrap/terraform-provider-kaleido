// Copyright 2018 Kaleido, a ConsenSys business

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package kaleido

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	kaleido "github.com/kaleido-io/kaleido-sdk-go/kaleido"
)

func TestKaleidoEnvironmentResource(t *testing.T) {
	consortium := kaleido.NewConsortium("terraformConsortEnv", "terraforming", "single-org")
	environment := kaleido.NewEnvironment("terraEnv", "terraforming", "quorum", "raft")
	envResourceName := "kaleido_environment.basicEnv"
	consortiumResourceName := "kaleido_consortium.basic"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConsortiumDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_basic(&consortium, &environment),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(envResourceName, consortiumResourceName),
				),
			},
		},
	})
}

func TestKaleidoEnvironmentResourcePickRelease(t *testing.T) {
	consortium := kaleido.NewConsortium("oldie", "terraforming", "single-org")
	environment := kaleido.NewEnvironment("oldieEnv", "terraforming", "quorum", "raft")
	consortiumResourceName := "kaleido_consortium.oldie"
	envResourceName := "kaleido_environment.oldieEnv"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckConsortiumDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEnvironmentConfig_oldRelease(&consortium, &environment),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckEnvironmentExists(envResourceName, consortiumResourceName),
					resource.TestCheckResourceAttr("kaleido_environment.oldieEnv", "release_id", "u0qaonpmzz"),
				),
			},
		},
	})
}

func testAccEnvironmentConfig_basic(consortium *kaleido.Consortium, environ *kaleido.Environment) string {
	return fmt.Sprintf(`resource "kaleido_consortium" "basic" {
		name = "%s"
		description = "%s"
		mode = "%s"
		}
		resource "kaleido_environment" "basicEnv" {
			consortium_id = "${kaleido_consortium.basic.id}"
			name = "%s"
			description = "%s"
			env_type = "%s"
			consensus_type = "%s"
		}
		`, consortium.Name,
		consortium.Description,
		consortium.Mode,
		environ.Name,
		environ.Description,
		environ.Provider,
		environ.ConsensusType)
}

func testAccEnvironmentConfig_oldRelease(consortium *kaleido.Consortium, environ *kaleido.Environment) string {
	return fmt.Sprintf(`resource "kaleido_consortium" "oldie" {
		name = "%s"
		description = "%s"
		mode = "%s"
		}
		resource "kaleido_environment" "oldieEnv" {
			consortium_id = "${kaleido_consortium.oldie.id}"
			name = "%s"
			description = "%s"
			env_type = "%s"
			consensus_type = "%s"
			release_id = "u0qaonpmzz"
		}
		`, consortium.Name,
		consortium.Description,
		consortium.Mode,
		environ.Name,
		environ.Description,
		environ.Provider,
		environ.ConsensusType)
}

func testAccCheckEnvironmentExists(envResource, consortiumResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[envResource]

		if !ok {
			return fmt.Errorf("Not found: %s", envResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Not terraform resource instance for %s", envResource)
		}

		consortium, ok := s.RootModule().Resources[consortiumResource]
		if !ok {
			return fmt.Errorf("Not found: %s", envResource)
		}

		if consortium.Primary.ID == "" {
			return fmt.Errorf("Not terraform resource instance for %s", envResource)
		}

		envId := rs.Primary.Attributes["id"]
		if envId != rs.Primary.ID {
			return fmt.Errorf("Terraform id mismatch for environment %s and %s", envId, rs.Primary.ID)
		}

		client := testAccProvider.Meta().(kaleido.KaleidoClient)
		var environment kaleido.Environment
		res, err := client.GetEnvironment(consortium.Primary.ID, envId, &environment)

		if err != nil {
			return err
		}

		if res.StatusCode() != 200 {
			return fmt.Errorf("Expected environment with id %s, status was: %d.", envId, res.StatusCode())
		}

		return nil
	}
}
