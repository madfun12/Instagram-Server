# Week 1 Project

I kind of fumbled around with a few ideas and I think this one is the best first project for me, since it has pretty big work implications. I have a pretty big potential issue that I'll get into later that may delay this projects finish, but I'm hoping it will get resolved soon.

This project idea is pretty straightforward. I'm wanting to implement a new "cartridge" at work (not a fan of that term, I prefer component or widget) that essentially acts just like lightwidget or the default Instagram embed code, where there's a section on the storefront homepage that has some basic Instagram information like the profile name, picture, followers, and the 6-10 most recent posts.

I looked into potentially just making a call from the front end, but of course they have CORS blocked so it needs to be done server-side. So this project has a few stages.

-   The frontend
    This will likeley just be a placeholder using React as I'm not sure what the current capabilities of eSAT are. Essentially all it will be is an application that serves as a point to launch the authentication process, and then send the access information to the server. I haven't decided or thought about how we're going to deal with actually displaying the instagram content to the end user yet because I think that would involve some sort of authentication (although while I'm typing this I think we may just have a search bar to open up the instagram feed of the supplied username just for easiness).

-   The server
    Once the front end provides the server with the access token from Meta, we'll store it in a database and tie it to a username, along with an expiration time. The server will have a few endpoint open through a REST api:
    /POST: Post a new user with an access token and an expiration date for the access token
    /GET: Retrieve a user's information. Name, pfp, follower count, page URL, posts, post URLs, etc.
    /PATCH: Update a user's access token and expiration date every so often to keep the auth up to date.

*   Idea for week 2 - a data visualization of the prime gap frequency distribution

Tomorrow - Thursday March 7 2024

-   Need to learn sql so that I can store the access token, userid, and expiration of the account in the DB

-   Set it up in these steps
    -   The code should come from the front end. That can be exchanged for a short-lived-access-token
    -   Short lived token can be exchanged for a long term access token
