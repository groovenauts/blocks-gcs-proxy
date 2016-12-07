require 'spec_helper'

describe Magellan::Gcs::Proxy::GCP do
  describe :project_id do
    let(:project_id) { 'dummy-project-id' }

    before { Magellan::Gcs::Proxy::GCP.reset }
    context 'local' do
      it 'with $BLOCKS_BATCH_PROJECT_ID' do
        backup = ENV['BLOCKS_BATCH_PROJECT_ID']
        ENV['BLOCKS_BATCH_PROJECT_ID'] = project_id
        begin
          expect(Magellan::Gcs::Proxy::GCP.project_id).to eq project_id
        ensure
          ENV['BLOCKS_BATCH_PROJECT_ID'] = backup
        end
      end

      it 'without $BLOCKS_BATCH_PROJECT_ID' do
        backup = ENV['BLOCKS_BATCH_PROJECT_ID']
        ENV['BLOCKS_BATCH_PROJECT_ID'] = nil
        begin
          expect do
            Magellan::Gcs::Proxy::GCP.project_id
          end.to raise_error(SocketError)
        ensure
          ENV['BLOCKS_BATCH_PROJECT_ID'] = backup
        end
      end
    end

    context 'on GKE or GCE' do
      let(:header) { { 'Metadata-Flavor' => 'Google' }.freeze }
      around do |example|
        backup = ENV['BLOCKS_BATCH_PROJECT_ID']
        ENV['BLOCKS_BATCH_PROJECT_ID'] = nil
        begin
          example.run
        ensure
          ENV['BLOCKS_BATCH_PROJECT_ID'] = backup
        end
      end

      let(:res) { double(:res) }
      it 'valid' do
        require 'net/http'
        expect(res).to receive(:code).and_return('200')
        expect(res).to receive(:body).and_return(project_id)
        expect_any_instance_of(Net::HTTP).to receive(:get).with('/computeMetadata/v1/project/project-id', header).and_return(res)
        expect(Magellan::Gcs::Proxy::GCP.project_id).to eq project_id
      end

      it 'invalid' do
        require 'net/http'
        expect(res).to receive(:code).and_return('400').twice
        expect(res).to receive(:body).and_return('Something wrong!')
        expect_any_instance_of(Net::HTTP).to receive(:get).with('/computeMetadata/v1/project/project-id', header).and_return(res)
        expect do
          Magellan::Gcs::Proxy::GCP.project_id
        end.to raise_error '[400] Something wrong!'
      end
    end
  end
end
