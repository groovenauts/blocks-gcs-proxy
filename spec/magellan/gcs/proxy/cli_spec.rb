require "spec_helper"

describe Magellan::Gcs::Proxy::Cli do
  context :case1 do
    let(:template) do
      "cmd1 %{download_files.foo} %{download_files.bar} %{attrs.baz} %{uploads_dir} %{attrs.qux}"
    end
    let(:downloads_dir){ '/tmp/workspace/downloads' }
    let(:uploads_dir){ '/tmp/workspace/uploads' }

    let(:download_files) do
      {
        'foo' => 'gs://bucket1/path/to/foo',
        'bar' => 'gs://bucket1/path/to/bar',
      }
    end
    let(:local_download_files) do
      {
        'foo' => "#{downloads_dir}/path/to/foo",
        'bar' => "#{downloads_dir}/path/to/bar",
      }
    end
    let(:upload_files) do
      [
        'gs://bucket1/path/to/file1',
        'gs://bucket1/path/to/file2',
        'gs://bucket1/path/to/file3',
      ]
    end

    let(:msg) do
      attrs = {
        'download_files' => download_files,
        'baz' => 60,
        'qux' => 'data1 data2 data3',
        'upload_files' => upload_files,
      }
      double(:msg, attributes: attrs)
    end

    let(:context) do
      Magellan::Gcs::Proxy::Context.new('/tmp/workspace', download_files, upload_files)
    end

    subject{ Magellan::Gcs::Proxy::Cli.new(template) }
    it :build_command do
      r = subject.build_command(msg, context)
      expected = 'cmd1 /tmp/workspace/downloads/path/to/foo /tmp/workspace/downloads/path/to/bar 60 /tmp/workspace/uploads data1 data2 data3'
      expect(r).to eq expected
    end
  end

  context :case2 do
    let(:template) do
      "cmd2 %{attrs.foo} %{download_files.bar} %{uploads_dir} %{download_files.baz} %{download_files.qux}"
    end
    let(:downloads_dir){ '/tmp/workspace/downloads' }
    let(:uploads_dir){ '/tmp/workspace/uploads' }

    let(:download_files) do
      {
        'bar' => 'gs://bucket2/path/to/bar',
        'baz' => 'gs://bucket2/path/to/baz',
        'qux' => [
          'gs://bucket2/path/to/qux1',
          'gs://bucket2/path/to/qux2',
        ],
      }
    end
    let(:upload_files) do
      [
        'gs://bucket2/path/to/file1',
        'gs://bucket2/path/to/file2',
        'gs://bucket2/path/to/file3',
      ]
    end

    let(:msg) do
      attrs = {
        'foo' => 123,
        'download_files' => download_files,
        'upload_files' => upload_files,
      }
      double(:msg, attributes: attrs)
    end

    let(:context) do
      Magellan::Gcs::Proxy::Context.new('/tmp/workspace', download_files, upload_files)
    end

    subject{ Magellan::Gcs::Proxy::Cli.new(template) }
    it :build_command do
      r = subject.build_command(msg, context)
      expected = 'cmd2 123 /tmp/workspace/downloads/path/to/bar /tmp/workspace/uploads /tmp/workspace/downloads/path/to/baz /tmp/workspace/downloads/path/to/qux1 /tmp/workspace/downloads/path/to/qux2'
      expect(r).to eq expected
    end
  end

end
