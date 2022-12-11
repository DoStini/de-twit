import {registerPostsUpdate, retrieveTimeline} from "../../services/posts";
import {addNewPost, postsStore, userPostsStore} from "../../actions/posts";
import {searchUser} from "../../services/users";

interface Params {
    [user: string]: string
}

interface PageLoad {
    [params: string]: Params
}

export async function load({ params }: PageLoad) {
    try {
        const userData = await searchUser(params.user);
        userPostsStore.set(userData.posts)
        return {
            user: userData
        }
    } catch (e) {
        console.error(e)
        postsStore.set([])
        return {
            error: e
        }
    }
}
