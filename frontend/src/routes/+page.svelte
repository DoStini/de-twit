<script lang="ts">
    import "../app.postcss";
    import Posts from "../components/post/Posts.svelte";
    import Fab from "../components/Fab.svelte";
    import {addPost, postsStore} from "../actions/posts.js"
    import type PostData from "../types/PostData.js";

    let posts: PostData[]
    postsStore.subscribe((value) => posts = value)

</script>

<Posts posts={posts}/>

<Fab action={() => addPost({
    text: "Ola mundo",
    username: "andremoreira9",
    timestamp: new Date(Date.now())
})}/>
