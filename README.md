# accounts

`accounts` is a subsystem of [Lockbox](https://lockbox.dev). Its responsibility
is to keep track of different ways a user can log in to your service.

When logging in, a user may be identified by an email address, Google account,
public key, or other identifier. The `accounts` module are how these
identifiers get associated with a "user" in your system.

## Design Goals

`accounts` is meant to be a discrete subsystem in the overall
[Lockbox](https://lockbox.dev) system. It tries to have clear boundaries of
responsibility and limit its responsibilities to only the things that it is
uniquely situated to do. Like all Lockbox subsystems, `accounts` is meant to be
an interchangeable part of the system, easily replaceable. All of its
functionality should be exposed through its API, instead of relying on other
subsystems importing it directly.

`accounts` tries to enable a user-friendly authentication experience. This
means allowing users to register multiple ways to authenticate with a service
and use them interchangeably, rather than asking users to remember which method
they used when registering.

## Implementation

`accounts` is implemented largely as a datastore and access mechanism for an
`Account` type. `Account` types just tie an `Account` ID to to a profile ID.
Profile IDs are expected to be controlled by `accounts`--it will decide what
new users' profile IDs are, and will not offload that responsibility to another
system.

The IDs for `Account`s must be unique, even across different authentication
methods. They are expected to be the same value a user would specify when
choosing a login method--an email address, a username, etc. The intended user
experience is that a user would enter their `Account` ID in an input box for
email-based and username-based authentication methods, or click a separate
button for an OAuth or OpenID login experience.

For email and username logins, the email or username as entered should be used
as the `Account` ID--`accounts` will make sure to use case-insensitive
uniqueness constraints and comparisons.

For OAuth and OpenID authentication methods, use whatever the authentication
provider uses as an account ID, unless an email address is available, in which
case, use that as the `Account` ID, instead. This allows users to log in using
an email-based flow or an OAuth or OpenID flow interchangeably.

## Scope

`accounts` is solely responsible for managing the connection between a user and
their various ways of logging into the system.

The questions `accounts` is meant to answer for the system include:

  * Which user is authenticating?
  * How can a specific user authenticate?
  * How does a user add a new way to authenticate?

The things `accounts` is explicitly not expected to do include:

  * Actually authenticate users. `accounts` is meant to tie an authentication
    method to a user, not actually do the authentication.
  * Managing ACLs.
  * Managing user profile information. Applications have specific enough needs
    for profiles that `accounts` can't reasonably abstract them. Instead,
    accounts just stores the profile ID for a user, which applications can then
    use to retrieve the profile information.

## Repository Structure

The base directory of the repository is used to set the logical framework and
shared types that will be used to talk about the subsystem. This largely means
defining types and interfaces.

The storers directory contains a collection of implementations of the `Storer`
interface, each in their own package. These packages should only have unit
tests, if any tests. The `Storer` acceptance tests in `storer_test.go` have
common acceptance testing for `Storer` implementations, and all `Storer`
implementations in the storers directory should register their tests there. If
the tests have setup requirements like databases or credentials, the tests
should only register themselves if these credentials are found.

The apiv1 directory contains the first version of the API interface. Breaking
changes should be published in a separate apiv2 package, so that both versions
of the API can be run simultaneously.
