require_relative "helper"

describe "patch operations" do
  let(:author) { "zypper-docker test suite" }
  let(:message) { "this is a test" }
  let(:ssl_bug) { "911399" }
  let(:ruby_bug) { "907809" }
  let(:ssl_cve) { "CVE-2014-3570" }
  let(:ruby_cve) { "CVE-2014-9130" }
  let(:ssl_patch) { "openSUSE-2014-671" }
  let(:ruby_patch) { "openSUSE-2015-6" }

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

  context 'listing patches' do
    it 'show patch name' do
      output = Cheetah.run(
        "zypper-docker", "lp", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include(ssl_patch)
      expect(output).to include(ruby_patch)
    end

    it 'can show the bugzilla number' do
      output = Cheetah.run(
        "zypper-docker", "lp", "--bugzilla", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include(ssl_bug)
      expect(output).to include(ruby_bug)
    end

    it 'can filter by bugzilla number' do
      output = Cheetah.run(
        "zypper-docker", "lp", "--bugzilla=#{ssl_bug}", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include(ssl_bug)
      expect(output).not_to include(ruby_bug)
    end

    it 'can show the cve number' do
      output = Cheetah.run(
        "zypper-docker", "lp", "--cve", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include(ssl_cve)
      expect(output).to include(ruby_cve)
    end

    it 'can filter by cve number' do
      output = Cheetah.run(
        "zypper-docker", "lp", "--cve=#{ssl_cve}", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include(ssl_cve)
      expect(output).not_to include(ruby_cve)
    end

    it 'can filter by date' do
      output = Cheetah.run(
        "zypper-docker", "lp", "--date", "2013-12-1", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include('No updates found')

      output = Cheetah.run(
        "zypper-docker", "lp", "--date", "2015-12-1", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include("openSUSE-2014-697") # another ssl issue
    end

    it 'can show the issue type' do
      output = Cheetah.run(
        "zypper-docker", "lp", "--issues", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include('bugzilla')
      expect(output).to include('cve')
    end

    it 'can filter by issue type' do
      output = Cheetah.run(
        "zypper-docker", "lp", "--issues=cve", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).not_to include('bugzilla')
      expect(output).to include('cve')
    end

    it 'filter by category name' do
      output = Cheetah.run(
        "zypper-docker", "lp", "--category", "security", Settings::VULNERABLE_IMAGE,
        stdout: :capture)
      expect(output).to include(ruby_patch)
    end

    context 'apply patches' do
      before(:each) do
        if docker_image_exists?(@patched_image_repo, @patched_image_tag)
          remove_docker_image(@patched_image)
        end
      end

      it 'can apply by bugzilla number' do
        Cheetah.run(
          "zypper-docker", "patch",
          "--author", author,
          "--message", message,
          "--bugzilla=#{ssl_bug}",
          Settings::VULNERABLE_IMAGE,
          @patched_image)
        expect(docker_image_exists?(@patched_image_repo, @patched_image_tag)).to be true

        output = Cheetah.run(
          "zypper-docker", "lp", "--bugzilla=#{ssl_bug}", @patched_image,
          stdout: :capture)
        expect(output).not_to include(ssl_bug)

        check_commit_details(author, message, @patched_image)
        expect(docker_inspect(@patched_image, ".Config.Entrypoint")).to eq "{[]}"
        expect(docker_inspect(@patched_image, ".Config.Cmd")).to eq "{[/bin/sh -c]}"
      end

      it 'can apply by cve number' do
        Cheetah.run(
          "zypper-docker", "patch",
          "--author", author,
          "--message", message,
          "--cve=#{ssl_cve}",
          Settings::VULNERABLE_IMAGE,
          @patched_image)
        expect(docker_image_exists?(@patched_image_repo, @patched_image_tag)).to be true

        output = Cheetah.run(
          "zypper-docker", "lp", "--cve=#{ssl_cve}", @patched_image,
          stdout: :capture)
        expect(output).not_to include(ssl_cve)

        check_commit_details(author, message, @patched_image)
        expect(docker_inspect(@patched_image, ".Config.Entrypoint")).to eq "{[]}"
        expect(docker_inspect(@patched_image, ".Config.Cmd")).to eq "{[/bin/sh -c]}"
      end

      it 'can apply by date' do
        Cheetah.run(
          "zypper-docker", "patch",
          "--author", author,
          "--message", message,
          "--date", "2015-2-1",
          Settings::VULNERABLE_IMAGE,
          @patched_image)
        expect(docker_image_exists?(@patched_image_repo, @patched_image_tag)).to be true

        output = Cheetah.run(
          "zypper-docker", "lp", "--date", "2015-8-1", @patched_image,
          stdout: :capture)
        expect(output).not_to include(ruby_patch)

        check_commit_details(author, message, @patched_image)
        expect(docker_inspect(@patched_image, ".Config.Entrypoint")).to eq "{[]}"
        expect(docker_inspect(@patched_image, ".Config.Cmd")).to eq "{[/bin/sh -c]}"
      end

      it 'apply by category name' do
        Cheetah.run(
          "zypper-docker", "patch",
          "--author", author,
          "--message", message,
          "--category", "security",
          Settings::VULNERABLE_IMAGE,
          @patched_image)
        expect(docker_image_exists?(@patched_image_repo, @patched_image_tag)).to be true

        output = Cheetah.run(
          "zypper-docker", "lp", @patched_image,
          stdout: :capture)
        expect(output).not_to include('security')
        expect(output).to include('recommended')

        check_commit_details(author, message, @patched_image)
        expect(docker_inspect(@patched_image, ".Config.Entrypoint")).to eq "{[]}"
        expect(docker_inspect(@patched_image, ".Config.Cmd")).to eq "{[/bin/sh -c]}"
      end

      it "does not overwrite the contents of the cmd and entrypoint" do
        @image_tag = "1.0"
        @image = "#{Settings::ENTRY_CMD_IMAGE_REPO}:#{@image_tag}"

        if docker_image_exists?(Settings::ENTRY_CMD_IMAGE_REPO, @image_tag)
          remove_docker_image(@image)
        end

        Cheetah.run(
          "zypper-docker", "patch",
          "--author", author,
          "--message", message,
          Settings::ENTRY_CMD_IMAGE,
          @image)

        expect(docker_image_exists?(Settings::ENTRY_CMD_IMAGE_REPO, @image_tag)).to be true

        check_commit_details(author, message, @image)
        expect(docker_inspect(@image, ".Config.Entrypoint")).to eq "{[cat]}"
        expect(docker_inspect(@image, ".Config.Cmd")).to eq "{[/etc/os-release]}"
      end

      it "can run zypper on a non-root image and reset the user afterwards" do
        @image_tag = "1.0"
        @image = "#{Settings::NORMAL_USER_IMAGE_REPO}:#{@image_tag}"

        if docker_image_exists?(Settings::NORMAL_USER_IMAGE_REPO, @image_tag)
          remove_docker_image(@image)
        end

        Cheetah.run(
          "zypper-docker", "patch",
          "--author", author,
          "--message", message,
          Settings::NORMAL_USER_IMAGE,
          @image)

        expect(docker_image_exists?(Settings::NORMAL_USER_IMAGE_REPO, @image_tag)).to be true

        check_commit_details(author, message, @image)
        expect(docker_inspect(@image, ".Config.User")).to eq "1337:1337"
      end
    end

    it 'checks patches' do
      exception = nil

      begin
        Cheetah.run(
          "zypper-docker", "pchk", Settings::VULNERABLE_IMAGE)
      rescue Cheetah::ExecutionFailed => e
        exception = e
      end
      expect(exception).not_to be_nil
      expect(exception.status.exitstatus).to eq(101)
      expect(exception.stdout).to include('security patches')
      expect(exception.stdout).to include('patches needed')
    end
  end

  context "analyze a running container" do
    before :all do
      @keep_alpine = docker_image_exists?("alpine", "latest")
      pull_image("alpine:latest") unless @keep_alpine

      @vul_container           = unique_name("vulnerable_container")
      @patched_container       = unique_name("patched_container")
      @not_suse_container      = unique_name("not_suse_container")
      @containers_to_terminate = []

      start_background_container(Settings::VULNERABLE_IMAGE, @vul_container)
      @containers_to_terminate << @vul_container

      if !docker_image_exists?(@patched_image_repo, @patched_image_tag)
        Cheetah.run("zypper-docker", "patch", "--category", "security",
          Settings::VULNERABLE_IMAGE, @patched_image)
      end
      expect(docker_image_exists?(@patched_image_repo, @patched_image_tag)).to be true
      start_background_container(@patched_image, @patched_container)
      @containers_to_terminate << @patched_container

      start_background_container("alpine:latest", @not_suse_container)
      @containers_to_terminate << @not_suse_container
    end

    after :all do
      @containers_to_terminate.each do |container|
        kill_and_remove_container(container)
      end

      remove_docker_image("alpine:latest") unless @keep_alpine
    end

    it "finds the pending updates of a SUSE-based image" do
      output = Cheetah.run("zypper-docker", "lpc", @vul_container, stdout: :capture)
      expect(output).to include("recommended")
    end

    it "does not find updates for patched containers" do
      output = Cheetah.run("zypper-docker", "lpc", @patched_container, stdout: :capture)
      expect(output).not_to include("security")
    end

    it "reports non-SUSE containers" do
      exception = nil

      begin
        output = Cheetah.run("zypper-docker", "lpc", @not_suse_container, stdout: :capture)
        expect(output).to include("alpine:latest which is not a SUSE system")
      rescue Cheetah::ExecutionFailed => e
        exception = e
      end
      expect(exception).not_to be_nil
      expect(exception.status.exitstatus).to eq(1)
    end
  end

end
