terraform {
  backend "s3" {
    bucket         = "terraform-state-storage-586877430255"
    dynamodb_table = "terraform-state-lock-586877430255"
    region         = "us-west-2"

    // THIS MUST BE UNIQUE
    key = "smee.tfstate"
  }
}

provider "aws" {
  region = "us-west-2"
}

data "aws_ssm_parameter" "eks_cluster_endpoint" {
  name = "/eks/av-cluster-endpoint"
}

provider "kubernetes" {
  host        = data.aws_ssm_parameter.eks_cluster_endpoint.value
  config_path = "~/.kube/config"
}

data "aws_ssm_parameter" "pg_username" {
  name = "/rds/av-main/smee_username"
}

data "aws_ssm_parameter" "pg_password" {
  name = "/rds/av-main/smee_password"
}

data "aws_ssm_parameter" "pg_hostname" {
  name = "/rds/av-main/hostname"
}

data "aws_ssm_parameter" "pg_port" {
  name = "/rds/av-main/port"
}

data "aws_ssm_parameter" "client_id" {
  name = "/env/smee/client-id"
}

data "aws_ssm_parameter" "client_secret" {
  name = "/env/smee/client-secret"
}

data "aws_ssm_parameter" "redis_url" {
  name = "/env/smee/redis-url"
}

data "aws_ssm_parameter" "hub_address" {
  name = "/env/hub-address"
}

module "smee" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "smee"
  image          = "docker.pkg.github.com/byuoitav/smee/smee-dev"
  image_version  = "60b5d53"
  container_port = 8080
  repo_url       = "https://github.com/byuoitav/smee"

  // optional
  image_pull_secret = "github-docker-registry"
  public_urls       = ["newsmee.av.byu.edu"]
  container_env = {
    "GIN_MODE" = "release"
  }
  container_args = [
    "--port", "8080",
    "--log-level", "info",
    "--hub-url", "ws://${data.aws_ssm_parameter.hub_address.value}",
    "--client-id", data.aws_ssm_parameter.client_id.value,
    "--client-secret", data.aws_ssm_parameter.client_secret.value,
    "--redis-url", data.aws_ssm_parameter.redis_url.value,
    "--postgres-url", "postgres://${data.aws_ssm_parameter.pg_username.value}:${data.aws_ssm_parameter.pg_password.value}@${data.aws_ssm_parameter.pg_hostname.value}:${data.aws_ssm_parameter.pg_port.value}/av?pool_max_conns=4",
  ]
  health_check = false
}
