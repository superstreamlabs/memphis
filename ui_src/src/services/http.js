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

import { message } from 'antd';
import axios from 'axios';

import { SERVER_URL, SHOWABLE_ERROR_STATUS_CODE, AUTHENTICATION_ERROR_STATUS_CODE } from '../config';
import { ApiEndpoints } from '../const/apiEndpoints';
import { LOCAL_STORAGE_TOKEN, LOCAL_STORAGE_EXPIRED_TOKEN } from '../const/localStorageConsts.js';
import AuthService from './auth';

export async function httpRequest(method, endPointUrl, data = {}, headers = {}, queryParams = {}, authNeeded = true, timeout = 0) {
    if (authNeeded) {
        const token = localStorage.getItem(LOCAL_STORAGE_TOKEN);
        headers['Authorization'] = 'Bearer ' + token;
    }
    const HTTP = axios.create({
        withCredentials: true
    });
    if (method !== 'GET' && method !== 'POST' && method !== 'PUT' && method !== 'DELETE')
        throw {
            status: 400,
            message: `Invalid HTTP method`,
            data: { method, endPointUrl, data }
        };
    try {
        const url = `${SERVER_URL}${endPointUrl}`;
        const res = await HTTP({
            method,
            url,
            data,
            headers,
            timeout,
            params: queryParams
        });
        const results = res.data;
        return results;
    } catch (err) {
        if (endPointUrl !== ApiEndpoints.LOGIN && err?.response?.status === AUTHENTICATION_ERROR_STATUS_CODE) {
            localStorage.clear();
            window.location.assign('/login');
        }
        if (err?.response?.data?.message !== undefined && err?.response?.status === SHOWABLE_ERROR_STATUS_CODE) {
            message.warning({
                key: 'memphisWarningMessage',
                content: err?.response?.data?.message,
                duration: 5,
                style: { cursor: 'pointer' },
                onClick: () => message.destroy('memphisWarningMessage')
            });
        }
        if (err?.response?.data?.message !== undefined && err?.response?.status === 500) {
            message.error({
                key: 'memphisErrorMessage',
                content: 'We have some issues. Please contact support.',
                duration: 5,
                style: { cursor: 'pointer' },
                onClick: () => message.destroy('memphisErrorMessage')
            });
        }
        throw err.response;
    }
}

export async function handleRefreshTokenRequest() {
    const HTTP = axios.create({
        withCredentials: true
    });
    try {
        const url = `${SERVER_URL}${ApiEndpoints.REFRESH_TOKEN}`;
        const res = await HTTP({ method: 'POST', url });
        const now = new Date();
        const expiryToken = now.getTime() + res.data.expires_in;
        if (process.env.REACT_APP_SANDBOX_ENV) {
            localStorage.setItem(LOCAL_STORAGE_TOKEN, res.data.jwt);
            localStorage.setItem(LOCAL_STORAGE_EXPIRED_TOKEN, expiryToken);
        } else {
            await AuthService.saveToLocalStorage(res.data);
        }
        return true;
    } catch (err) {
        localStorage.clear();
        window.location.assign('/login');
        return false;
    }
}
