import {registerPostsUpdate, retrieveTimeline} from "../services/posts";
import {addNewPost, postsStore} from "../actions/posts";

export async function load() {
    try {
        const posts = await retrieveTimeline()
        console.log(posts)
        postsStore.set(posts)
        return {}
    } catch (e) {
        console.error(e)
        postsStore.set([])
        return {}
    }
}
