require "magellan/gcs/proxy"

module Magellan
  module Gcs
    module Proxy
      module ExpandVariable

        class InvalidReferenceError < StandardError
        end

        module_function

        def dig_variables(variable_ref, data)
          vars = variable_ref.split(".").map{|i| (/\A\d+\z/.match(i)) ? i.to_i : i }
          value = vars.inject(data) do |tmp, v|
            case v
            when String
              if tmp.respond_to?(:[]) && tmp.respond_to?(:include?)
                if tmp.include?(v)
                  tmp[v]
                else
                  raise InvalidReferenceError, variable_ref
                end
              else
                raise InvalidReferenceError, variable_ref
              end
            when Integer
              case tmp
              when Array
                if tmp.size > v
                  tmp[v]
                else
                  raise InvalidReferenceError, variable_ref
                end
              else
                raise InvalidReferenceError, variable_ref
              end
            end
          end
        end

        def expand_variables(str, data, quote_string: false)
          data ||= {}
          str.gsub(/\%\{\s*([\w.]+)\s*\}/) do |m|
            var = Regexp.last_match(1)
            value =
              begin
                dig_variables(var, data)
              rescue InvalidReferenceError
                ""
              end

            case value
            when String then quote_string ? value.to_json : value
            else value.to_json
            end
          end
        end
      end

    end
  end
end
