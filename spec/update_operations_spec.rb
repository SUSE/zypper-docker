require_relative "helper"

describe "update operations" do
  let(:author) { "zypper-docker test suite" }
  let(:message) { "this is a test" }

  before :all do
    @patched_image_repo = "zypper-docker-patched-image"
    @patched_image_tag  = "1.0"
    @patched_image      = "#{@patched_image_repo}:#{@patched_image_tag}"
  end

  after :all do
    if docker_image_exists?(@patched_image_repo, @patched_image_tag)
      remove_docker_image(@patched_image)
    end
  end

  it "lists updates" do
    output = Cheetah.run("zypper-docker", "lu", Settings::VULNERABLE_IMAGE, stdout: :capture)
    expect(output).to include("alsa-utils")
  end

  it "applies updates" do
    Cheetah.run(
      "zypper-docker", "up",
      "--author", author,
      "--message", message,
      Settings::VULNERABLE_IMAGE,
      @patched_image)

    expect(docker_image_exists?(@patched_image_repo, @patched_image_tag)).to be true
    output = Cheetah.run("zypper-docker", "lu", @patched_image, stdout: :capture)
    expect(output).not_to include("alsa-utils")

    check_commit_details(author, message, @patched_image)
  end

  it "refuses to overwrite an existing image while doing an update" do
    expect(docker_image_exists?(@patched_image_repo, @patched_image_tag)).to be true

    expect {
      Cheetah.run(
        "zypper-docker", "up",
        "--author", author,
        "--message", message,
        Settings::VULNERABLE_IMAGE,
      @patched_image)
    }.to raise_error(Cheetah::ExecutionFailed)
  end

end
