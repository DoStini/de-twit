import type PostData from "./PostData";

type UserData = {
    username: string,
    posts: PostData[],
    following: boolean
}

export default UserData;
