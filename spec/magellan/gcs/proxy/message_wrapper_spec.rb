require 'spec_helper'

describe Magellan::Gcs::Proxy::MessageWrapper do
  context :case1 do
    let(:download_files) do
      {
        'foo' => 'gs://bucket1/path/to/foo',
        'bar' => 'gs://bucket1/path/to/bar'
      }
    end
    let(:upload_files) do
      [
        'gs://bucket1/path/to/file1',
        'gs://bucket1/path/to/file2',
        'gs://bucket1/path/to/file3'
      ]
    end
    let(:msg) do
      attrs = {
        'download_files' => download_files.to_json,
        'baz' => 60,
        'qux' => 'data1 data2 data3',
        'upload_files' => upload_files.to_json,
      }
      double(:msg, attributes: attrs)
    end
    let(:context) do
      Magellan::Gcs::Proxy::Context.new(msg).tap do |c|
        allow(c).to receive(:workspace).and_return('/tmp/workspace')
      end
    end

    context 'wrapper' do
      subject { Magellan::Gcs::Proxy::MessageWrapper.new(context) }
      it { expect(subject['downloads_dir']).to eq '/tmp/workspace/downloads' }
      it { expect(subject['uploads_dir']).to eq '/tmp/workspace/uploads' }
      it { expect(subject['attrs']).to be_a(Magellan::Gcs::Proxy::MessageWrapper::Attrs) }
    end

    context 'attrs' do
      subject { Magellan::Gcs::Proxy::MessageWrapper.new(context)['attrs'] }
      it { expect(subject['download_files']).to eq download_files }
      it { expect(subject['upload_files']).to eq upload_files }
      it { expect(subject['baz']).to eq 60 }
      it { expect(subject['qux']).to eq 'data1 data2 data3' }
    end
  end

  context :case2 do
    let(:download_files) do
      {
        'bar' => 'gs://bucket2/path/to/bar',
        'baz' => 'gs://bucket2/path/to/baz',
        'qux' => [
          'gs://bucket2/path/to/qux1',
          'gs://bucket2/path/to/qux2'
        ]
      }
    end
    let(:upload_files) do
      [
        'gs://bucket2/path/to/file1',
        'gs://bucket2/path/to/file2',
        'gs://bucket2/path/to/file3'
      ]
    end

    let(:msg) do
      attrs = {
        'foo' => 123,
        'download_files' => download_files.to_json,
        'upload_files' => upload_files.to_json,
      }
      double(:msg, attributes: attrs)
    end
    let(:context) do
      Magellan::Gcs::Proxy::Context.new(msg).tap do |c|
        allow(c).to receive(:workspace).and_return('/tmp/workspace')
      end
    end

    context 'attrs' do
      subject { Magellan::Gcs::Proxy::MessageWrapper.new(context)['attrs'] }
      it { expect(subject['foo']).to eq 123 }
      it { expect(subject['download_files']).to eq download_files }
      it { expect(subject['upload_files']).to eq upload_files }
    end
  end
end
