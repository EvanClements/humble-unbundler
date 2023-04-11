## To Do:

- [x] Remove sensitive parameters and hardcoded user info
- [x] Create system to get needed user info from config.json
- [ ] What info is needed? "_simpleauth_sess" cookie?
- [ ] Upload to GitHub once config system setup
- [ ] Create a README.md with a description, how to set up, and how to use
- [ ] Web UI or Terminal UI?

### Progress:

- [x] Use cookie to access Humble Bundle user info
- [x] Use cookie to access further info for each bundle
- [ ] Add information to SQlite db
- [ ] Download Direct Download Links
- [ ] Create UI

### Minimum Effective Product?

The minimum needed would be to gather the purchases, 
download all the direct download links (ebook, audio, apks, etc.),
and list games and their statuses (redeemed or not redeemed?)

That would be v1. v2 would include some way to check 
if the game is already in the user's steam/origin/uplay 
library, and a way to redeem the unredeemed keys.

For v3, that would be further polishing: a UI and 
a Docker image (or at least a Dockerfile), and 
automated binary releases.
