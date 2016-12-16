require 'magellan/gcs/proxy'
require 'magellan/gcs/proxy/log'

module Magellan
  module Gcs
    module Proxy
      class PubsubSubscription
        attr_reader :name, :delay
        def initialize(name, delay: 1)
          @name = name
          @delay = delay
        end

        def listen
          loop do
            if msg = wait_for_message
              yield msg
            else
              sleep delay
            end
          end
        end

        def pull_req
          @pull_req ||= Google::Apis::PubsubV1::PullRequest.new(max_messages: 1, return_immediately: true)
        end

        def wait_for_message
          # #<Google::Apis::PubsubV1::ReceivedMessage:0x007fdc440b58d8
          #   @ack_id="...",
          #   @message=#<Google::Apis::PubsubV1::Message:0x007fdc440be140
          #     @attributes={"download_files"=>"[\"gs://bucket1/path/to/file1\"]"},
          #     @message_id="50414480536440",
          #     @publish_time="2016-12-17T08:08:35.599Z">>
          res = GCP.pubsub.pull_subscription(name, pull_req)
          msg = (res.received_messages || []).first
          msg.nil? ? nil : MessageWrapper.new(self, msg)
        end

        class MessageWrapper
          attr_reader :subscription, :original

          # @param [Magellan::Gcs::Proxy::PubsubSubscription] subscription
          # @param [Google::Apis::PubsubV1::ReceivedMessage] original
          def initialize(subscription, original)
            @subscription = subscription
            @original = original
          end

          def method_missing(name, *args, &block)
            receiver =
              case name
              when :message_id, :attributes, :publish_time, :data then original.message
              when :ack_id then original
              end
            receiver ? receiver.send(name, *args, &block) : super
          end

          def respond_to_missing?(name)
            case name
            when :message_id, :attributes, :publish_time, :data then true
            when :ack_id then true
            else false
            end
          end

          def ack_options
            # https://github.com/google/google-api-ruby-client/blob/master/lib/google/apis/options.rb#L48
            @ack_options ||= Google::Apis::RequestOptions.new(authorization: GCP.auth, retries: 5)
          end

          def acknowledge!
            req = Google::Apis::PubsubV1::AcknowledgeRequest.new(ack_ids: [ack_id])
            GCP.pubsub.acknowledge_subscription(subscription.name, req, options: ack_options)
          end
          alias :ack! :acknowledge!

          def delay!(new_deadline)
            req = Google::Apis::PubsubV1::ModifyAckDeadlineRequest.new(ack_deadline_seconds: new_deadline, ack_ids: [ack_id])
            GCP.pubsub.modify_subscription_ack_deadline(subscription.name, req)
          end
        end
      end
    end
  end
end
