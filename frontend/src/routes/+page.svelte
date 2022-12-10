<script lang="ts">
    import "../app.postcss";
    import Posts from "../components/post/Posts.svelte";
    import Fab from "../components/Fab.svelte";
    import { addNewPost, postsStore } from "../actions/posts.js"
    import type PostData from "../types/PostData.js";
    import {newPostsStore} from "../actions/posts";

    let posts: PostData[]
    postsStore.subscribe((value) => posts = value)

    let newPosts: PostData[]
    newPostsStore.subscribe((value) => newPosts = value)
</script>

<Posts posts={posts} newPosts={newPosts}/>

<Fab action={() => addNewPost({
    text: "Ola mundo",
    username: "andremoreira9",
    timestamp: new Date(Date.now())
})}/>
