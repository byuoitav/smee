terraform {
  backend "s3" {
    bucket         = "terraform-state-storage-887007127029"
    dynamodb_table = "terraform-state-lock-887007127029"
    region         = "us-west-2"

    // THIS MUST BE UNIQUE
    key = "smee-dev.tfstate"
  }
}

provider "aws" {
  region = "us-west-2"
}

data "aws_ssm_parameter" "eks_cluster_endpoint" {
  name = "/eks/av-dev-cluster-endpoint"
}

provider "kubernetes" {
  host        = data.aws_ssm_parameter.eks_cluster_endpoint.value
  config_path = "~/.kube/config"
}

data "aws_ssm_parameter" "pg_username" {
  name = "/rds/av-dev/smee_username"
}

data "aws_ssm_parameter" "pg_password" {
  name = "/rds/av-dev/smee_password"
}

data "aws_ssm_parameter" "pg_hostname" {
  name = "/rds/av-dev/hostname"
}

data "aws_ssm_parameter" "pg_port" {
  name = "/rds/av-dev/port"
}

data "aws_ssm_parameter" "client_id" {
  name = "/env/smee-dev/client-id"
}

data "aws_ssm_parameter" "client_secret" {
  name = "/env/smee-dev/client-secret"
}

data "aws_ssm_parameter" "redis_url" {
  name = "/env/smee-dev/redis-url"
}

data "aws_ssm_parameter" "hub_address" {
  name = "/env/hub-address"
}

data "aws_ssm_parameter" "gateway_url" {
  name = "/env/smee-dev/gateway-url"
}

data "aws_ssm_parameter" "redirect_url" {
  name = "/env/smee-dev/redirect-url"
}

data "aws_ssm_parameter" "opa_url" {
  name = "/env/smee-dev/opa-url"
}

data "aws_ssm_parameter" "opa_token" {
  name = "/env/smee-dev/opa-token"
}

data "aws_ssm_parameter" "command_server" {
  name = "/env/smee-dev/command-server"
}

data "aws_ssm_parameter" "command_token" {
  name = "/env/smee-dev/command-token"
}

data "aws_ssm_parameter" "couch_address" {
  name = "/env/couch-address"
}

data "aws_ssm_parameter" "couch_username" {
  name = "/env/couch-username"
}

data "aws_ssm_parameter" "couch_password" {
  name = "/env/couch-password"
}

module "smee" {
  source = "github.com/byuoitav/terraform//modules/kubernetes-deployment"

  // required
  name           = "smee"
  image          = "docker.pkg.github.com/byuoitav/smee/smee-dev"
  image_version  = "fe470f9"
  container_port = 8080
  repo_url       = "https://github.com/byuoitav/smee"

  // optional
  image_pull_secret = "github-docker-registry"
  public_urls       = ["monitoring.avdev.byu.edu"]
  container_env = {
    "GIN_MODE" = "release"
  }
  container_args = [
    "--port", "8080",
    "--log-level", "info",
    "--web-root", "/website",
    "--hub-url", "ws://${data.aws_ssm_parameter.hub_address.value}",
    "--client-id", data.aws_ssm_parameter.client_id.value,
    "--client-secret", data.aws_ssm_parameter.client_secret.value,
    "--redis-url", data.aws_ssm_parameter.redis_url.value,
    "--postgres-url", "postgres://${data.aws_ssm_parameter.pg_username.value}:${data.aws_ssm_parameter.pg_password.value}@${data.aws_ssm_parameter.pg_hostname.value}:${data.aws_ssm_parameter.pg_port.value}/av?pool_max_conns=4",
    "--gateway-url", data.aws_ssm_parameter.gateway_url.value,
    "--redirect-url", data.aws_ssm_parameter.redirect_url.value,
    "--opa-url", data.aws_ssm_parameter.opa_url.value,
    "--opa-token", data.aws_ssm_parameter.opa_token.value,
    "--command-server", data.aws_ssm_parameter.command_server.value,
    "--command-token", data.aws_ssm_parameter.command_token.value,
    "--couch-address", data.aws_ssm_parameter.couch_address.value,
    "--couch-username", data.aws_ssm_parameter.couch_username.value,
    "--couch-password", data.aws_ssm_parameter.couch_password.value
  ]
  health_check = false
}