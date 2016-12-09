require 'magellan/gcs/proxy'

require 'logger'
require 'json'

module Magellan
  module Gcs
    module Proxy
      class PubsubSustainer
        class << self
          def run(message)
            raise "#{name}.run requires block" unless block_given?
            if c = Proxy.config[:sustainer]
              t = Thread.new(message, c['delay'], c['interval']) do |msg, delay, interval|
                Thread.current[:processing_message] = true
                new(msg, delay: delay, interval: interval).run
              end
              begin
                yield
              ensure
                t[:processing_message] = false
                t.join
              end
            else
              yield
            end
          end
        end

        attr_reader :message, :delay, :interval
        def initialize(message, delay: 10, interval: nil)
          @message = message
          @delay = delay.to_i
          @interval = (interval || @delay * 0.9).to_f
        end

        def run
          loop do
            sleep(interval)
            break unless Thread.current[:processing_message]
            message.delay! delay
          end
        end
      end
    end
  end
end
