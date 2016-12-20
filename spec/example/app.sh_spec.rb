require 'spec_helper'

require 'fileutils'
require 'tmpdir'

describe "example/app.sh" do
  let(:command_path) { File.expand_path('../../../example/app.sh', __FILE__) }
  let(:test_file_path) { 'bucket1/foo/bar/baz.txt' }
  let(:test_suffix) { 'test' }
  let(:test_content) { 'Lorem Ipsum' }
  let(:tmpdir) { @tmpdir ||= Dir.mktmpdir }
  let(:downloads_dir) { File.join(tmpdir, 'downloads') }
  let(:uploads_dir) { File.join(tmpdir, 'uploads') }

  before do
    FileUtils.mkdir_p(downloads_dir)
    FileUtils.mkdir_p(uploads_dir)
  end
  after {  FileUtils.remove_entry_secure tmpdir }

  context "valid" do
    it do
      download_file = File.join(downloads_dir, test_file_path)
      FileUtils.mkdir_p(File.dirname(download_file))
      open(download_file, 'w') { |f| f.puts(test_content) }
      r = system([command_path, download_file, downloads_dir, uploads_dir, test_suffix].join(' '))
      expect(r).to be_truthy
      result_path = File.join(uploads_dir, test_file_path.sub(%r{\.txt\z}, "-#{test_suffix}.txt"))
      expect(File.read(result_path).strip).to eq test_content
    end
  end
end
