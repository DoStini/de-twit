import type PostData from "../types/PostData";

export const createPost: (post: PostData) => (void) = async (post: PostData) => {
}

export const retrieveTimeline : () => (Promise<PostData[]>) = async () => {
    return await Promise.resolve(
        [
            {username: "andremoreira9", text: "Awesome work guys!", timestamp: new Date()},
            {username: "marga", text: "Great! I'm currently merging the timelines", timestamp: new Date()},
            {username: "nuno", text: "Hi guys, I'm doing a massive refactor!", timestamp: new Date()},
            {username: "andremoreira9", text: "Weird sandals indeed", timestamp: new Date()},
            {username: "andremoreira9", text: "Weird sandals indeed 2", timestamp: new Date()},
            {username: "andremoreira9", text: "Weird sandals indeed 3", timestamp: new Date()},
        ]
    )
}
