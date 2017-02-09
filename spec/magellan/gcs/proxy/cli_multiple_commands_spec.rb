require 'spec_helper'

describe Magellan::Gcs::Proxy::Cli do
  let(:notification_topic_name) { 'progress-topic-for-rspec' }
  let(:config_data) do
    {
      'progress_notification' => {
        'topic' => notification_topic_name,
      },
      'loggers' => [
        { 'type' => 'stdout' },
        { 'type' => 'cloud_logging', 'log_name' => 'cloud-logging-for-rspec' },
      ],
      'commands' => {
        'key1' => 'cmd1 %{download_files}',
        'key2' => 'cmd2 %{download_files.bar} %{uploads_dir} %{download_files.baz}',
      }
    }
  end
  before do
    # GCP
    Magellan::Gcs::Proxy.config.reset
    allow(Magellan::Gcs::Proxy.config).to receive(:load_file).and_return(config_data)
    allow(Magellan::Gcs::Proxy::GCP).to receive(:pubsub).and_return(pubsub)
    allow(Magellan::Gcs::Proxy::GCP).to receive(:storage).and_return(storage)
    allow(Magellan::Gcs::Proxy::GCP).to receive(:logging).and_return(logging)
    allow(logging).to receive(:resource).with('container', {}).and_return(logging_resource)
  end

  let(:pubsub) { double(:pubsub) }
  let(:storage) { double(:storage) }
  let(:logging) do
    double(:logging).tap do |l|
      allow(l).to receive(:write_entries).with(an_instance_of(Google::Cloud::Logging::Entry),
                                               an_instance_of(Hash))
    end
  end
  let(:logging_resource) { double(:logging_resource) }

  context :with_commands do
    let(:template){ '%{attrs.foo}' }
    let(:downloads_dir) { '/tmp/workspace/downloads' }
    let(:uploads_dir) { '/tmp/workspace/uploads' }

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

    subject { Magellan::Gcs::Proxy::Cli.new(template) }

    context "key1" do
      let(:msg) do
        attrs = {
          'foo' => 'key1',
          'download_files' => download_files.to_json,
          'upload_files' => upload_files,
        }
        double(:msg, attributes: attrs)
      end

      let(:context) do
        Magellan::Gcs::Proxy::Context.new(msg).tap do |c|
          allow(c).to receive(:workspace).and_return('/tmp/workspace')
        end
      end

      it do
        r = subject.build_command(context)
        expected = 'cmd1'\
                   ' /tmp/workspace/downloads/bucket2/path/to/bar'\
                   ' /tmp/workspace/downloads/bucket2/path/to/baz'\
                   ' /tmp/workspace/downloads/bucket2/path/to/qux1'\
                   ' /tmp/workspace/downloads/bucket2/path/to/qux2'
        expect(r).to eq expected
      end
    end

    context "key2" do
      let(:msg) do
        attrs = {
          'foo' => 'key2',
          'download_files' => download_files.to_json,
          'upload_files' => upload_files,
        }
        double(:msg, attributes: attrs)
      end

      let(:context) do
        Magellan::Gcs::Proxy::Context.new(msg).tap do |c|
          allow(c).to receive(:workspace).and_return('/tmp/workspace')
        end
      end

      it do
        r = subject.build_command(context)
        expected = 'cmd2'\
                   ' /tmp/workspace/downloads/bucket2/path/to/bar'\
                   ' /tmp/workspace/uploads'\
                   ' /tmp/workspace/downloads/bucket2/path/to/baz'
        expect(r).to eq expected
      end
    end

    context "invalid key" do
      let(:msg) do
        attrs = {
          'foo' => 'invalid-key',
          'download_files' => download_files.to_json,
          'upload_files' => upload_files,
        }
        double(:msg, attributes: attrs)
      end

      let(:context) do
        Magellan::Gcs::Proxy::Context.new(msg).tap do |c|
          allow(c).to receive(:workspace).and_return('/tmp/workspace')
        end
      end

      it do
        expect{
          subject.build_command(context)
        }.to raise_error(Magellan::Gcs::Proxy::BuildError)
      end
    end
  end

end
