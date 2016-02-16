Aerolith 3.0 will be written in Go and Javascript; with the main goals
of being multiplayer and mobile-ready.

## Authentication

We wish to migrate the database from Aerolith 2.0 as much as possible, in terms of users/passwords/etc but there is no need to handle registration.

Use an OpenID/similar provider; we will have Facebook and Google as options. That should cover the large majority of our use case. Allow username/password login for people who haven't migrated.

## Gameplay

### Single-player
- Daily challenges
- Study mode


### Multi-player
- Default to parallel solving
- Another mode for collaboratively clearing a wall
- max: 8+ players?

## Lists

- Allow list sharing
- Make mechanics a bit more intuitive

## Lexica

- Allow upload of custom lexica. 
- Go thread to make lexica files


## Flashcards
- For now keep Whitley cardz.. would need to rewrite the couple of backend views.
