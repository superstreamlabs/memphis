import {LOCAL_STORAGE_USER_ID} from "../const/localStorageConsts";
export function isCurrentUser(userId) {
    const currentUserId = localStorage.getItem(LOCAL_STORAGE_USER_ID);
    return parseInt(userId) === parseInt(currentUserId);
}