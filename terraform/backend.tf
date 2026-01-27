terraform {
  backend "s3" {
    bucket       = "shin-yokohama-event-notifier-tfstate"
    key          = "terraform.tfstate"
    region       = "ap-northeast-1"
    encrypt      = true
    use_lockfile = true
  }
}
