require_relative "helper"

describe "ps operations" do
  let(:author) { "zypper-docker test suite" }
  let(:message) { "this is a test" }

  before :all do
    @keep_alpine = docker_image_exists?("alpine", "latest")
    if !@keep_alpine
      pull_image("alpine:latest")
    end
    @patched_image_repo = "zypper-docker-patched-image"
    @patched_image_tag  = "1.0"
    @patched_image      = "#{@patched_image_repo}:#{@patched_image_tag}"

    @vul_container = "vulnerable_container"
    @not_suse_container = "not_suse_container"

    Cheetah.run(
      "docker", "run",
      "-d",
      "--name", @vul_container,
      Settings::VULNERABLE_IMAGE,
      "sleep", "1h")
    Cheetah.run(
      "docker", "run",
      "-d",
      "--name", @not_suse_container,
      "alpine:latest",
      "sleep", "1h")
  end

  after :all do
    kill_and_remove_container(@vul_container)
    kill_and_remove_container(@not_suse_container)

    remove_docker_image("alpine:latest") unless @keep_alpine
    if docker_image_exists?(@patched_image_repo, @patched_image_tag)
      remove_docker_image(@patched_image)
    end
  end

  it "handle uknown containers" do
    output = Cheetah.run("zypper-docker", "ps", stdout: :capture)
    expect(output).to include("The following containers have an unknown state")
    expect(output).to include(Settings::VULNERABLE_IMAGE)
    expect(output).to include("alpine:latest")
  end

  it "recognizes vulnerable containers" do
    Cheetah.run(
      "zypper-docker", "patch",
      "--author", author,
      "--message", message,
      Settings::VULNERABLE_IMAGE,
      @patched_image)

    output = Cheetah.run("zypper-docker", "ps", stdout: :capture).split("\n")
    expect(output[0]).to include("Running containers whose images have been updated")
    expect(output[1]).to include(Settings::VULNERABLE_IMAGE)
  end

  def kill_and_remove_container(container)
    Cheetah.run("docker", "kill", container)
    Cheetah.run("docker", "rm", container)
  end

end
