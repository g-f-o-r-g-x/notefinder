=head1

Notefinder - Perl interface to Notefinder (note-taking application)

=cut

package Notefinder;
use strict;
use warnings;

our $VERSION = "0.1";

sub Hello {
	print "Hello from Notefinder!\n";
}

package Notefinder::Note;

sub Print {
	my $note = shift;
	for my $k (sort keys %$note) {
		print "$k => $note->{$k}\n";
	}
}

1;
