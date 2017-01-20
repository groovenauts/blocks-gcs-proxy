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

  context :case1 do
    let(:template) do
      'cmd1 %{download_files.foo} %{download_files.bar} %{attrs.baz} %{uploads_dir} %{attrs.qux}'
    end
    let(:downloads_dir) { '/tmp/workspace/downloads' }
    let(:uploads_dir) { '/tmp/workspace/uploads' }

    let(:download_file_paths) do
      {
        'foo' => 'path/to/foo',
        'bar' => 'path/to/bar',
      }
    end
    let(:download_files) do
      download_file_paths.each_with_object({}) { |(k, v), d| d[k] = "gs://#{bucket_name}/#{v}" }
    end
    let(:local_download_files) do
      download_file_paths.each_with_object({}) { |(k, v), d| d[k] = "#{downloads_dir}/#{v}" }
    end
    let(:bucket_name) { 'bucket1' }
    let(:upload_files) do
      [
        "gs://#{bucket_name}/path/to/file1",
        "gs://#{bucket_name}/path/to/file2",
        "gs://#{bucket_name}/path/to/file3",
      ]
    end

    let(:message_id) { '1234567890' }
    let(:msg) do
      attrs = {
        'download_files' => download_files.to_json,
        'baz' => 60,
        'qux' => 'data1 data2 data3',
        'upload_files' => upload_files,
      }
      double(:msg, message_id: message_id, attributes: attrs)
    end

    let(:context) do
      Magellan::Gcs::Proxy::Context.new(msg).tap do |c|
        allow(c).to receive(:workspace).and_return('/tmp/workspace')
      end
    end

    let(:cmd1_by_msg) do
      'cmd1'\
      ' /tmp/workspace/downloads/bucket1/path/to/foo'\
      ' /tmp/workspace/downloads/bucket1/path/to/bar'\
      ' 60 /tmp/workspace/uploads data1 data2 data3'
    end

    subject { Magellan::Gcs::Proxy::Cli.new(template) }
    it :build_command do
      r = subject.build_command(context)
      expect(r).to eq cmd1_by_msg
    end

    describe :process do
      let(:bucket) { double(:bucket) }
      let(:upload_file_path1) { 'path/to/upload_file1' }
      before do
        # Context
        expect(Magellan::Gcs::Proxy::Context).to receive(:new).with(msg).and_return(context)

        # Notification Topic
        allow(pubsub).to receive(:publish_topic).with(notification_topic_name, an_instance_of(Google::Apis::PubsubV1::PublishRequest)) do |_topic, req|
          expect(req.messages.length).to eq 1
          msg = req.messages.first
          expect(msg.attributes).to be_an(Hash)
        end

        # Download
        expect(storage).to receive(:bucket).with(bucket_name)
          .and_return(bucket).exactly(download_files.length).times
        download_file_paths.each_with_index do |(_key, path), idx|
          gcs_file = double(:"gcs_file_#{idx}")
          expect(bucket).to receive(:file).with(path).and_return(gcs_file)
          expect(gcs_file).to receive(:download).with("#{downloads_dir}/#{bucket_name}/#{path}")
        end

        # Execute
        any_composite_logger = an_instance_of(Magellan::Gcs::Proxy::CompositeLogger)
        expect(LoggerPipe).to receive(:run).with(any_composite_logger, cmd1_by_msg, returns: :none, logging: :both, dry_run: nil)

        # Upload
        expect(Dir).to receive(:chdir).with(uploads_dir).and_yield
        expect(Dir).to receive(:glob).with('*').and_yield(bucket_name)
        expect(Dir).to receive(:chdir).with(bucket_name).and_yield
        expect(Dir).to receive(:glob).with('**/*').and_yield(upload_file_path1)
        expect(context).to receive(:directory?).with(upload_file_path1).and_return(false)
        expect(storage).to receive(:bucket).with(bucket_name).and_return(bucket)
        expect(bucket).to receive(:create_file).with(upload_file_path1, upload_file_path1)

        # Ack
        expect(msg).to receive(:acknowledge!)
      end
      context 'normal' do
        it do
          subject.process(msg)
        end
      end
    end
  end

  context :case2 do
    let(:template) do
      'cmd2 %{attrs.foo} %{download_files.bar} %{uploads_dir} %{download_files.baz} %{download_files.qux}'
    end
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

    let(:msg) do
      attrs = {
        'foo' => 123,
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

    subject { Magellan::Gcs::Proxy::Cli.new(template) }
    it :build_command do
      r = subject.build_command(context)
      expected = 'cmd2 123'\
                 ' /tmp/workspace/downloads/bucket2/path/to/bar'\
                 ' /tmp/workspace/uploads'\
                 ' /tmp/workspace/downloads/bucket2/path/to/baz'\
                 ' /tmp/workspace/downloads/bucket2/path/to/qux1'\
                 ' /tmp/workspace/downloads/bucket2/path/to/qux2'
      expect(r).to eq expected
    end
  end
end
