## Database
Of course we will use PostgreSQL. CouchDB may also be used to save word lists and PouchDB on client side for syncing.

SQLite for the individual lexicon files. Why three databases? That's fine tho.

### Tables to migrate

Need to migrate a bunch of tables from Django, but there are a lot more that we don't need.

accounts_aerolithprofile - keep track of membership status, wordwalls style, etc
auth_user - our users.
base_lexicon - simple metadata about lexica
wordwalls_dailychallenge - all of these tables need to be moved in order to preserve history
wordwalls_dailychallengeleaderboard
wordwalls_dailychallengeleaderboardentry
wordwalls_dailychallengemissedbingos
wordwalls_namedlist - we have a script to generate these so maybe not as important
wordwalls_savedlist

### Tables to move elsewhere
base_alphagram - Move to SQLite.
base_word - Move to SQLite
blog_posts - what to do with these?
django_comments - move with the blog posts?

## Gameplay

We will use sockets (go-sockjs?) or perhaps Firebase for single player and multiplayer experience. 

If we use sockjs let's use PostgreSQL for the pub-sub -- Redis is great but too many moving parts right now.

There are positives and negatives to doing server-side or client-side validation.

### Server-side validation
+ Easier to sync multiplayer 
+ Easier to mitigate some forms of cheating - but cheating will always be too easy
+ Easier to keep track of solved words/etc for list saving
- Poor internet connection causes visible lag

### Client-side validation
+ More immune to lag/poor connections -- may be more important for mobile
- Easier to cheat if all it requires is a packet saying, x words solved in y seconds.

## APIs

The API should be something like JSON-RPC over websockets