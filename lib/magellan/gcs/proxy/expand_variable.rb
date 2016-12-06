require 'magellan/gcs/proxy'

module Magellan
  module Gcs
    module Proxy
      module ExpandVariable
        class InvalidReferenceError < StandardError
        end

        module_function

        def dig_variables(variable_ref, data)
          vars = variable_ref.split('.').map { |i| /\A\d+\z/ =~ i ? i.to_i : i }
          value = vars.inject(data) do |tmp, v|
            dig_variable(tmp, v, variable_ref)
          end
        end

        def dig_variable(tmp, v, variable_ref)
          case v
          when String
            if tmp.respond_to?(:[]) && tmp.respond_to?(:include?)
              return tmp[v] if tmp.include?(v)
            end
          when Integer
            case tmp
            when Array
              return tmp[v] if tmp.size > v
            end
          end
          raise InvalidReferenceError, variable_ref
        end

        def expand_variables(str, data, quote_string: false)
          data ||= {}
          str.gsub(/\%\{\s*([\w.]+)\s*\}/) do |_m|
            var = Regexp.last_match(1)
            value =
              begin
                dig_variables(var, data)
              rescue InvalidReferenceError
                ''
              end

            case value
            when String then quote_string ? value.to_s : value
            when Array then value.flatten.join(' ')
            else value.to_s
            end
          end
        end
      end
    end
  end
end
