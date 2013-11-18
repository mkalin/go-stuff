use strict;
use LWP;
use XML::XPath;

my $baseUrl = 'http://localhost:8080/cliches/sayings.jsp';
my $ua = LWP::UserAgent->new;

getTest($baseUrl);
## my $cmd = 'curl --request GET localhost:8080/cliches2/';
## system($cmd);

sub getTest() {
    my ($url) = @_;

    print $url, "\n";

    my $request = HTTP::Request->new(GET => $url);
    my $response = $ua->request($request);
    
    if ($response->is_success) {
	print "Raw XML:\n$response\n";
    }
    else {
	print "$response->status_line\n"
    }
}

