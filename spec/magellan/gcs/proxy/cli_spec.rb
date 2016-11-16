require "spec_helper"

describe Magellan::Gcs::Proxy::Cli do
  context :case1 do
    let(:template) do
      "cmd1 %{attrs.download_files.foo} %{attrs.download_files.bar} %{attrs.baz} %{uploads_dir} %{attrs.qux}"
    end
    let(:msg) do
      attrs = {
        'download_files' => {
          'foo' => 'gs://bucket1/path/to/foo',
          'bar' => 'gs://bucket1/path/to/bar',
        },
        'baz' => 60,
        'qux' => 'data1 data2 data3',
        'upload_files' => [
          'gs://bucket1/path/to/file1',
          'gs://bucket1/path/to/file2',
          'gs://bucket1/path/to/file3',
        ],
      }
      double(:msg, attributes: attrs)
    end
    let(:context) do
      {
        workspace: '/tmp/workspace',
        downloads_dir: '/tmp/workspace/downloads',
        uploads_dir: '/tmp/workspace/uploads',
      }
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
      "cmd2 %{attrs.foo} %{attrs.download_files.bar} %{uploads_dir} %{attrs.download_files.baz} %{attrs.download_files.qux}"
    end
    let(:msg) do
      attrs = {
        'foo' => 123,
        'download_files' => {
          'bar' => 'gs://bucket2/path/to/bar',
          'baz' => 'gs://bucket2/path/to/baz',
          'qux' => [
            'gs://bucket2/path/to/qux1',
            'gs://bucket2/path/to/qux2',
          ],
        },
        'upload_files' => [
          'gs://bucket2/path/to/file1',
          'gs://bucket2/path/to/file2',
          'gs://bucket2/path/to/file3',
        ]
      }
      double(:msg, attributes: attrs)
    end
    let(:context) do
      {
        workspace: '/tmp/workspace',
        downloads_dir: '/tmp/workspace/downloads',
        uploads_dir: '/tmp/workspace/uploads',
      }
    end

    subject{ Magellan::Gcs::Proxy::Cli.new(template) }
    it :build_command do
      r = subject.build_command(msg, context)
      expected = 'cmd2 123 /tmp/workspace/downloads/path/to/bar /tmp/workspace/uploads /tmp/workspace/downloads/path/to/baz /tmp/workspace/downloads/path/to/qux1 /tmp/workspace/downloads/path/to/qux2'
      expect(r).to eq expected
    end
  end

end
