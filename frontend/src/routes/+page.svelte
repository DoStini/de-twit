<script lang="ts">
    import "../app.postcss";
    import Posts from "../components/post/Posts.svelte";
    import Fab from "../components/Fab.svelte";
    import type PostData from "../types/PostData.js";

    const posts: PostData[] = [
        {
            username: "andremoreira9",
            text: "A beterraba caiu do céu! Finalmente :) \n Portugal ganhou! alé alé alé!",
            timestamp: new Date(Date.now())
        },
        {
            username: "andremoreira9",
            text: "A beterraba caiu do céu! Finalmente :) \n Portugal ganhou! alé alé alé!",
            timestamp: new Date(Date.now())
        },
        {
            username: "nuno",
            text: "A beterraba caiu do céu! Finalmente :) \n Portugal ganhou! alé alé alé!",
            timestamp: new Date(Date.now())
        },
        {
            username: "marga",
            text: "A beterraba caiu do céu! Finalmente :) \n Portugal ganhou! alé alé alé!",
            timestamp: new Date(Date.now())
        },
        {
            username: "andremoreira9",
            text: "A beterraba caiu do céu! Finalmente :) \n Portugal ganhou! alé alé alé!",
            timestamp: new Date(Date.now())
        },
        {
            username: "marga",
            text: "A beterraba caiu do céu! Finalmente :) \n Portugal ganhou! alé alé alé!",
            timestamp: new Date(Date.now())
        },
        {
            username: "marga",
            text: "A beterraba caiu do céu! Finalmente :) \n Portugal ganhou! alé alé alé!",
            timestamp: new Date(Date.now())
        },
    ];
</script>

<Posts posts={posts}/>

<Fab/>
