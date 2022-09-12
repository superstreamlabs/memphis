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

const environment = process.env.NODE_ENV ? process.env.NODE_ENV : 'development';
const SERVER_URL_PRODUCTION = `${window.location.href.split('//')[1].split('/')[0]}/api`;
const SSL_PREFIX = window.location.protocol === 'https:' ? 'https://' : 'http://';

export const SERVER_URL = environment === 'production' ? `${SSL_PREFIX}${SERVER_URL_PRODUCTION}` : 'http://localhost:9000/api';

export const HANDLE_REFRESH_INTERVAL = 600000;
export const SHOWABLE_ERROR_STATUS_CODE = 666;
export const AUTHENTICATION_ERROR_STATUS_CODE = 401;
export const DOC_URL = 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/memphis/overview';
export const PRIVACY_URL = 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/memphis/privacy';
export const SOCKET_URL = environment === 'production' ? `${SSL_PREFIX}${SERVER_URL_PRODUCTION}` : 'localhost:9000/api';

export const GOOGLE_CLIENT_ID = '916272522459-u0f4n2lh9llsielb3l5rob3dnt1fco76.apps.googleusercontent.com';
export const REDIRECT_URI = environment === 'production' ? 'https://sandbox.memphis.dev/login' : 'http://localhost:9000/login';
export const GITHUB_CLIENT_ID = environment === 'production' ? '4dc1b3238c4d7563e426' : '51b0330eb3b34bc8f641';
