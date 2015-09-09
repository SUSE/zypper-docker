require "rspec/core/rake_task"
require "ci/reporter/rake/rspec"

desc "Run the integration tests"
RSpec::Core::RakeTask.new("spec")

desc "Run tests and generate JUnit compatible files"
task test: ["ci:setup:rspec", "spec"]
