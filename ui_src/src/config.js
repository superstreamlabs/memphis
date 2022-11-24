// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server

const environment = process.env.NODE_ENV ? process.env.NODE_ENV : 'development';
const SERVER_URL_PRODUCTION = `${window.location.href.split('//')[1].split('/')[0]}/api`;
var WS_SERVER_URL_PRODUCTION = `${window.location.href.split('//')[1].split('/')[0]}`;
if (WS_SERVER_URL_PRODUCTION.includes(':'))
    // for urls contain port
    WS_SERVER_URL_PRODUCTION = WS_SERVER_URL_PRODUCTION.split(':')[0];
const SSL_PREFIX = window.location.protocol === 'https:' ? 'https://' : 'http://';

export const SERVER_URL = environment === 'production' ? `${SSL_PREFIX}${SERVER_URL_PRODUCTION}` : 'http://localhost:9000/api';
const WS_PREFIX = window.location.href.includes('https') ? 'wss' : 'ws';
export const URL = window.location.href;

export const HANDLE_REFRESH_INTERVAL = 600000;
export const SHOWABLE_ERROR_STATUS_CODE = 666;
export const AUTHENTICATION_ERROR_STATUS_CODE = 401;
export const DOC_URL = 'https://docs.memphis.dev/memphis/memphis/overview';
export const PRIVACY_URL = 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/memphis/privacy';
export const SOCKET_URL = environment === 'production' ? `${WS_PREFIX}://${WS_SERVER_URL_PRODUCTION}:7770` : `${WS_PREFIX}://localhost:7770`;
export const GOOGLE_CLIENT_ID = '916272522459-u0f4n2lh9llsielb3l5rob3dnt1fco76.apps.googleusercontent.com';
export const REDIRECT_URI = environment === 'production' ? 'https://sandbox.memphis.dev/login' : 'http://localhost:9000/login';
export const GITHUB_CLIENT_ID = environment === 'production' ? '4dc1b3238c4d7563e426' : '51b0330eb3b34bc8f641';
export const CONNECT_APP_VIDEO = 'https://www.youtube.com/watch?v=-5YmxYRQsdw&t=3s';
export const CONNECT_CLI_VIDEO = 'https://www.youtube.com/watch?v=awXwaU4rBBQ&t=56s';
