# accounts

The `accounts` package encapsulates the part of the [auth system](https://impractical.co/auth) that maps the method a user logged in with to the identifier used in our systems to represent that user.

Put another way, it helps tie usernames, email addresses, Google accounts, and any future ways to login we may want to support into a single logical unit that we can think of as "a user".

## Implementation

Accounts consist of an ID, a profile ID, and some metadata. The ID is used to uniquely identify the account. We don't use a compound ID (e.g., a unique combination is of a type and ID, like "email" and "paddy@impractical.co") because one of the benefits we want from this system is for users to be able to log in with their email address or username interchangeably. To avoid confusion, it'd be best if they could enter either in a single input and the system just does The Right Thing. To do this, however, usernames and emails must be unique even when stored together.

The profile ID is a UUID, and is meaningless in the system. Its value is opaque. The only important properties it has is that it is immutable and consistent across all accounts for a shared "user".

### Emails and Usernames

Emails and usernames should be stored in the account field exactly as entered; a case-insensitive comparison will be used to match them at login time. Any input with an @ in it will be considered an email address, and so we'll try to send mail to it.

### Google ID Tokens

One of the ways to log into the system is using a [Google ID token](https://developers.google.com/identity/sign-in/web/backend-auth). These tokens will be validated and verified, then parsed into the user information they represent. That information provides the user's email, which will be used as the account ID as though the user had entered it themselves, got the email, and successfully entered the code/clicked the link.

## Scope

`accounts` is solely responsible for managing the connection between a user and their various ways of logging into the system. The HTTP handlers it provides are responsible for verifying the authentication and authorization of the requests made against it, which will be coming from untrusted sources.

The questions `accounts` is meant to answer for the system include:

  * Which user is authenticating?
  * How can a specific user authenticate?
  * How does a user add a new way to authenticate?

The things `accounts` is explicitly not expected to do include:

  * Actually authenticate users. That is handled by the [`grants` package](httpps://impractical.co/auth/grants).
  * Sending any emails.
  * Managing ACLs.
  * Managing user profile information.
