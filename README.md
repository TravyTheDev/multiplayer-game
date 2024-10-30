### Kick Him in the Middle

This is an online fighting game made with a Go server and Ebitengine.

![](https://raw.githubusercontent.com/TravyTheDev/personal-site/refs/heads/main/public/images/kick-him.gif)

Honestly this started because I thought it would be funny and I was showing someone it's not hard to move something on someone else's screen. 

I started off doing this in Javascript but found Ebitengine and went from there.

### Optimizations

There's a lot that needs to be done.

I have no idea how to do menus in the game engine or how menus in general even work in games. Is it all React Native? If I can get menus working it's basically good to host and let people play. 

There is no score keeping. 

The "animations" are just showing a sprite depending on game state. I don't know if that's the correct way to do it.

The players fly because I couldn't figure out how to make characters "jump."

### Lessons Learned

I could have written my own way to send data, but I'd never have progressed so I just used JSON. 

Get someone with art skills to do art.

Game state has to be updated for both players, obvious, I know. 

Colliders are impossible when being hit from multiple angles.