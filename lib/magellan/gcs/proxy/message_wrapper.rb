require 'magellan/gcs/proxy'

module Magellan
  module Gcs
    module Proxy
      class MessageWrapper
        attr_reader :msg, :context
        def initialize(context)
          @msg = context.message
          @context = context
        end

        def [](key)
          case key.to_sym
          when :attrs, :attributes then return attributes
          when :data then return msg.data
          end
          context[key.to_sym]
        end

        def include?(key)
          k = key.to_sym
          context.include?(k) || [:attrs, :attributes, :data].include?(k)
        end

        def attributes
          Attrs.new(msg.attributes)
        end

        class Attrs
          attr_reader :data
          def initialize(data)
            @data = data
          end

          def [](key)
            value = data[key]
            if value.is_a?(String) && value =~ /\A\[.*\]\z|\A\{.*\}\z/
              begin
                JSON.parse(value)
              rescue
                value
              end
            else
              value
            end
          end

          def include?(key)
            data.include?(key) || data.include?(key.to_sym)
          end
        end
      end
    end
  end
end
