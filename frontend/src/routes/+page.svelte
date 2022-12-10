<script lang="ts">
    import "../app.postcss";
    import Fab from "../components/Fab.svelte";
    import { postsStore } from "../actions/posts.js"
    import type PostData from "../types/PostData.js";
    import {newPostsStore} from "../actions/posts";
    import {onMount} from "svelte";
    import {retrieveTimeline} from "../services/posts";
    import NewPostModal from "../components/post/NewPostModal.svelte";
    import {closeNewPostModal, openNewPostModal} from "../actions/newPostModal.js";
    import {createPost} from "../services/posts.js";
    import {env} from "$env/dynamic/public";
    import Posts from "../components/post/Posts.svelte";

    let posts: PostData[]
    postsStore.subscribe((value) => posts = value)

    let newPosts: PostData[]
    newPostsStore.subscribe((value) => newPosts = value)

    let loading: boolean = false

    let handleCreatePost = async (data) => {
        loading = true
        try {
            await createPost({
                username: env.PUBLIC_USERNAME,
                text: data.content,
                timestamp: new Date()
            });
            closeNewPostModal()
            loading = false
            return true
        } catch (e) {
            console.error(e)
            loading = false
            return true
        }
    }

    onMount(async () => {
        try {
            const posts = await retrieveTimeline()
            postsStore.set(posts)
        } catch (e) {

        }
    })
</script>

<Posts posts={posts} newPosts={newPosts}/>

<Fab action={openNewPostModal}/>

<NewPostModal loading={loading} close={closeNewPostModal} submit={handleCreatePost}/>
