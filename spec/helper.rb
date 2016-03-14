require "cheetah"
require "pathname"

class Settings
  BASE_IMAGE_REPO = "opensuse"
  BASE_IMAGE_TAG  = "13.2"
  BASE_IMAGE      = "#{BASE_IMAGE_REPO}:#{BASE_IMAGE_TAG}"

  VULNERABLE_IMAGE_REPO = "zypper-docker-tests-vulnerable-image"
  VULNERABLE_IMAGE_TAG = "0.1"
  VULNERABLE_IMAGE = "#{VULNERABLE_IMAGE_REPO}:#{VULNERABLE_IMAGE_TAG}"

  ENTRY_CMD_IMAGE_REPO = "zypper-docker-tests-entrypoint-cmd-image"
  ENTRY_CMD_IMAGE_TAG = "0.1"
  ENTRY_CMD_IMAGE = "#{ENTRY_CMD_IMAGE_REPO}:#{ENTRY_CMD_IMAGE_TAG}"

  NORMAL_USER_IMAGE_REPO = "zypper-docker-tests-normal-user-image"
  NORMAL_USER_IMAGE_TAG = "0.1"
  NORMAL_USER_IMAGE = "#{NORMAL_USER_IMAGE_REPO}:#{NORMAL_USER_IMAGE_TAG}"
end

module SpecHelper
  def docker_image_exists?(repo, tag)
    if tag == "" || tag.nil?
      tag = "latest"
    end
    output = Cheetah.run("docker", "images", repo, stdout: :capture)
    output.include?(tag)
  end

  def pull_image(image)
    system("docker pull #{image}")
  end

  def container_running?(container)
    output = Cheetah.run("docker", "ps", stdout: :capture)
    output.include?(container)
  end

  def kill_and_remove_container(container)
    Cheetah.run("docker", "kill", container) if container_running?(container)
    Cheetah.run("docker", "rm", "-f", container)
  rescue Cheetah::ExecutionFailed => e
    puts "Error while running: #{e.commands}"
    puts "Stdout: #{e.stdout}"
    puts "Stderr: #{e.stderr}"
  end

  # Make sure that the given image exists. The name parameter corresponds to
  # the name of the final image with the tag already included in it. The
  # dockerfile corresponds to the name of the dockerfile. The path to this
  # dockerfile is assumed to be inside of the "docker" directory of the root of
  # this project. The dockerfile has to be based on BASE_IMAGE.
  def ensure_image_exists(name, dockerfile)
    # force pull of latest release of the opensuse:13.2 image
    system("docker pull #{Settings::BASE_IMAGE}")

    # Do not use cheetah, we want live streaming of what is happening
    puts "Building #{name}"
    cmd = "docker build" \
      " -f #{File.join(Pathname.new(__FILE__).parent.parent, "docker/#{dockerfile}")}" \
      " -t #{name}" \
      " #{Pathname.new(__FILE__).parent.parent}"
    puts cmd
    system(cmd)
    exit(1) if $? != 0
  end

  def ensure_vulnerable_image_exists
    ensure_image_exists(Settings::VULNERABLE_IMAGE, "Dockerfile-vulnerable-image")
  end

  def ensure_entrypoint_cmd_image_exists
    ensure_image_exists(Settings::ENTRY_CMD_IMAGE, "Dockerfile-entrypoint-and-cmd-set")
  end

  def ensure_normal_user_image_exists
    ensure_image_exists(Settings::NORMAL_USER_IMAGE, "Dockerfile-normal-user")
  end

  def remove_docker_image(image)
    Cheetah.run("docker", "rmi", "-f", image)
  end

  # Inspect the given docker image with the given format.
  def docker_inspect(image, format)
    res = Cheetah.run("docker", "inspect", "--format='{{#{format}}}'", image, stdout: :capture)
    return nil if res == "<nil>\n"
    res.strip
  end

  def docker_image_commit_details(image)
    [docker_inspect(image, ".Author"), docker_inspect(image, ".Comment")]
  end

  def check_commit_details(expected_author, expected_message, image)
    actual_author, actual_message = docker_image_commit_details(image)
    expect(expected_author.chomp).to eq(actual_author.chomp)
    expect(expected_message.chomp).to eq(actual_message.chomp)
  end

  # Makes the given name unique by adding the current time to it.
  def unique_name(name)
    "#{name}_#{Time.now.to_i}"
  end

  # Start a container running in the background doing a NOP (a simple sleep)
  def start_background_container(image, container_name, solve_conflicts=true)
    Cheetah.run(
      "docker", "run",
      "-d",
      "--entrypoint", "env",
      "--name", container_name,
      image,
      "sh", "-c", "sleep 1h")
  end

end

RSpec.configure do |config|
  config.include(SpecHelper)

  config.before :all do
    ensure_vulnerable_image_exists
    ensure_entrypoint_cmd_image_exists
    ensure_normal_user_image_exists
  end
end
