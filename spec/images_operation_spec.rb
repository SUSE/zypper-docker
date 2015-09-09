describe "update operations" do
  before :all do
    # hash with key: image id, bool set to true if the image is already there
    @image_status = {}

    ["busybox", "opensuse"].each do |image|
      available = docker_image_exists?(image, "latest")
      pull_image(image) unless available
      expect(docker_image_exists?(image, "latest")).to be true
      @image_status[image] = available
    end

  end

  after :all do
    @image_status.each do |image, to_remove|
      remove_docker_image(image) if to_remove
    end
  end

  it 'lists images that are based on either openSUSE or SLE' do
    output = Cheetah.run("zypper-docker", "-f", "images", stdout: :capture)

    expect(output).to include('opensuse')
    expect(output).not_to include('busybox')
  end

end

