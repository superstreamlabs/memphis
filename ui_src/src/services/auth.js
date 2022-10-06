// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import {
    LOCAL_STORAGE_ALREADY_LOGGED_IN,
    LOCAL_STORAGE_AVATAR_ID,
    LOCAL_STORAGE_CREATION_DATE,
    LOCAL_STORAGE_TOKEN,
    LOCAL_STORAGE_EXPIRED_TOKEN,
    LOCAL_STORAGE_USER_ID,
    LOCAL_STORAGE_USER_NAME,
    LOCAL_STORAGE_USER_TYPE,
    LOCAL_STORAGE_ALLOW_ANALYTICS,
    LOCAL_STORAGE_ENV,
    LOCAL_STORAGE_NAMESPACE,
    LOCAL_STORAGE_WELCOME_MESSAGE,
    LOCAL_STORAGE_FULL_NAME,
    LOCAL_STORAGE_SKIP_GET_STARTED
} from '../const/localStorageConsts';
import pathDomains from '../router';

const AuthService = (function () {
    const saveToLocalStorage = (userData) => {
        const now = new Date();
        const expiryToken = now.getTime() + userData.expires_in;

        localStorage.setItem(LOCAL_STORAGE_ALREADY_LOGGED_IN, userData.already_logged_in);
        localStorage.setItem(LOCAL_STORAGE_AVATAR_ID, userData.avatar_id);
        localStorage.setItem(LOCAL_STORAGE_CREATION_DATE, userData.creation_date);
        localStorage.setItem(LOCAL_STORAGE_TOKEN, userData.jwt);
        localStorage.setItem(LOCAL_STORAGE_USER_ID, userData.user_id);
        localStorage.setItem(LOCAL_STORAGE_USER_NAME, userData.username);
        localStorage.setItem(LOCAL_STORAGE_FULL_NAME, userData.full_name);
        localStorage.setItem(LOCAL_STORAGE_USER_TYPE, userData.user_type);
        localStorage.setItem(LOCAL_STORAGE_EXPIRED_TOKEN, expiryToken);
        localStorage.setItem(LOCAL_STORAGE_ALLOW_ANALYTICS, userData.send_analytics);
        localStorage.setItem(LOCAL_STORAGE_ENV, userData.env);
        localStorage.setItem(LOCAL_STORAGE_NAMESPACE, userData.namespace);
        if (userData.already_logged_in === false) {
            localStorage.setItem(LOCAL_STORAGE_WELCOME_MESSAGE, true);
        }
    };

    const logout = () => {
        const isSkipGetStarted = localStorage.getItem(LOCAL_STORAGE_SKIP_GET_STARTED);
        localStorage.clear();
        localStorage.setItem(LOCAL_STORAGE_SKIP_GET_STARTED, isSkipGetStarted);
        window.location.assign(pathDomains.login);
    };

    const isValidToken = () => {
        const tokenExpiryTime = localStorage.getItem(LOCAL_STORAGE_EXPIRED_TOKEN);
        if (Date.now() <= tokenExpiryTime) {
            return true;
        } else {
            return false;
        }
    };

    return {
        saveToLocalStorage,
        logout,
        isValidToken
    };
})();
export default AuthService;
