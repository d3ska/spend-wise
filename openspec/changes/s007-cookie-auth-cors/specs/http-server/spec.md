## MODIFIED Requirements

### Requirement: Auth endpoint - OAuth callback
The server SHALL expose `POST /api/v1/auth/{provider}/callback` that exchanges an OAuth code for a user profile and sets an httpOnly session cookie. The response body SHALL contain only the user object.

#### Scenario: Successful callback
- **WHEN** `POST /api/v1/auth/google/callback` is called with body `{"code":"valid-code","state":"matching-state"}`
- **THEN** the response status SHALL be `200 OK`
- **AND** the response body SHALL contain `{"user":{...}}` without a `token` field
- **AND** a `Set-Cookie` header SHALL be present with the JWT in an httpOnly cookie
