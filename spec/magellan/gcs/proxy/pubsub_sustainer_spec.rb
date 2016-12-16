require 'spec_helper'

describe Magellan::Gcs::Proxy::PubsubSustainer do
  describe '.run' do
    context 'with sustainer' do
      let(:delay) { 2 }
      let(:config_data) do
        {
          'loggers' => [{ 'type' => 'stdout' }],
          'sustainer' => {
            'delay' => delay,
          },
        }
      end
      let(:msg) { double(:msg) }

      before do
        Magellan::Gcs::Proxy.config.reset
        allow(Magellan::Gcs::Proxy.config).to receive(:load_file).and_return(config_data)
      end

      it do
        expect(msg).to receive(:delay!).with(delay).exactly(2).times
        Timeout.timeout 6 do
          Magellan::Gcs::Proxy::PubsubSustainer.run(msg) do
            sleep(5)
          end
        end
      end

      let(:error_message) { 'Unexpected Error' }
      it 'logging error on delay!' do
        Thread.current[:processing_message] = true
        sustainer = Magellan::Gcs::Proxy::PubsubSustainer.new(msg, delay: delay)
        logger = sustainer.send(:logger)
        expect(logger).to receive(:error).with(instance_of(RuntimeError))
        expect(msg).to receive(:delay!).with(delay).and_raise(error_message)
        Timeout.timeout 4 do
          sustainer.run do
            sleep(3)
          end
        end
      end
    end

    context 'without sustainer' do
      let(:config_data) do
        {
          'loggers' => [{ 'type' => 'stdout' }],
        }
      end
      let(:msg) { double(:msg) }

      before do
        Magellan::Gcs::Proxy.config.reset
        allow(Magellan::Gcs::Proxy.config).to receive(:load_file).and_return(config_data)
      end

      it do
        expect(msg).not_to receive(:delay!)
        Timeout.timeout 6 do
          Magellan::Gcs::Proxy::PubsubSustainer.run(msg) do
            sleep(5)
          end
        end
      end
    end
  end

  describe :send_delay do
    let(:delay) { 3 }
    let(:interval) { 2 }
    let(:msg) { double(:msg) }

    subject { Magellan::Gcs::Proxy::PubsubSustainer.new(msg, delay: delay, interval: interval) }

    context 'on Google::Cloud::UnavailableError' do
      # E, [2016-12-15T08:51:31.381860 #1] ERROR -- : 14:
      #      {
      #        "created":"@1481791891.380648742","description":"Secure read failed",
      #        "file":"src/core/lib/security/transport/secure_endpoint.c","file_line":157,"grpc_status":14,
      #        "referenced_errors":[
      #          {"created":"@1481791891.380596379","description":"EOF","file":"src/core/lib/iomgr/tcp_posix.c","file_line":235}
      #        ]
      #      } (Google::Cloud::UnavailableError)
      it 'retries until limit' do
        cnt = 0
        expect(msg).to receive(:delay!) do
          cnt += 1
          raise Google::Cloud::UnavailableError, '{"description":"Secure read failed"}' if cnt < 3
        end.exactly(3).times
        expect(subject).to receive(:next_limit).and_return(Time.now.to_f + delay).twice
        subject.send_delay
      end

      it 'gives up retrying after the next_limit passes' do
        expect(msg).to receive(:delay!) \
                       .with(delay) \
                       .and_raise(Google::Cloud::UnavailableError.new('{"description":"Secure read failed"}')) \
                       .exactly(4).times
        expect(subject).to receive(:next_limit).and_return(Time.now.to_f + delay).exactly(4).times
        expect { subject.send_delay }.to raise_error(Google::Cloud::UnavailableError)
      end
    end
  end
end
