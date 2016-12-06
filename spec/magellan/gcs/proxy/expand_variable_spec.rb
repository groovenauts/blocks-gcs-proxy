require 'spec_helper'

describe Magellan::Gcs::Proxy::MessageWrapper do
  include Magellan::Gcs::Proxy::ExpandVariable

  context :case1 do
    let(:downloads_dir) { '/tmp/workspace/downloads' }
    let(:uploads_dir) { '/tmp/workspace/uploads' }

    let(:bucket1) { 'bucket1' }
    let(:download_files) do
      {
        'foo' => "gs://#{bucket1}/path/to/foo",
        'bar' => "gs://#{bucket1}/path/to/bar",
      }
    end
    let(:local_download_files) do
      {
        'foo' => "#{downloads_dir}/#{bucket1}/path/to/foo",
        'bar' => "#{downloads_dir}/#{bucket1}/path/to/bar",
      }
    end
    let(:upload_files) do
      [
        "gs://#{bucket1}/path/to/file1",
        "gs://#{bucket1}/path/to/file2",
        "gs://#{bucket1}/path/to/file3",
      ]
    end

    let(:msg) do
      attrs = {
        'download_files' => download_files.to_json,
        'baz' => 60,
        'qux' => 'data1 data2 data3',
        'upload_files' => upload_files,
      }
      double(:msg, attributes: attrs)
    end

    let(:context) do
      Magellan::Gcs::Proxy::Context.new(msg).tap do |c|
        allow(c).to receive(:workspace).and_return('/tmp/workspace')
      end
    end

    let(:data) { Magellan::Gcs::Proxy::MessageWrapper.new(context) }

    it { expect(expand_variables('%{uploads_dir}', data)).to eq uploads_dir }
    it { expect(expand_variables('%{attrs.baz}', data)).to eq '60' }
    it { expect(expand_variables('%{download_files.foo}', data)).to eq local_download_files['foo'] }
    it { expect(expand_variables('%{download_files.bar}', data)).to eq local_download_files['bar'] }
    it { expect(expand_variables('%{attrs.download_files.foo}', data)).to eq download_files['foo'] }
    it { expect(expand_variables('%{attrs.download_files.bar}', data)).to eq download_files['bar'] }
  end

  context :case2 do
    let(:downloads_dir) { '/tmp/workspace/downloads' }
    let(:uploads_dir) { '/tmp/workspace/uploads' }

    let(:bucket2) { 'bucket2' }
    let(:download_files) do
      {
        'bar' => "gs://#{bucket2}/path/to/bar",
        'baz' => "gs://#{bucket2}/path/to/baz",
        'qux' => [
          "gs://#{bucket2}/path/to/qux1",
          "gs://#{bucket2}/path/to/qux2",
        ],
      }
    end

    let(:local_download_files) do
      {
        'bar' => "#{downloads_dir}/#{bucket2}/path/to/bar",
        'baz' => "#{downloads_dir}/#{bucket2}/path/to/baz",
        'qux' => [
          "#{downloads_dir}/#{bucket2}/path/to/qux1",
          "#{downloads_dir}/#{bucket2}/path/to/qux2",
        ],
      }
    end

    let(:upload_files) do
      [
        "gs://#{bucket2}/path/to/file1",
        "gs://#{bucket2}/path/to/file2",
        "gs://#{bucket2}/path/to/file3",
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

    let(:data) { Magellan::Gcs::Proxy::MessageWrapper.new(context) }

    it { expect(expand_variables('%{attrs.foo}', data)).to eq '123' }
    it { expect(expand_variables('%{download_files.qux}', data)).to eq local_download_files['qux'].join(' ') }
    it { expect(expand_variables('%{attrs.download_files.qux}', data)).to eq download_files['qux'].join(' ') }
  end
end
