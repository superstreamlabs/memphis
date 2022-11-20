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
const SSL_PREFIX = window.location.protocol === 'https:' ? 'https://' : 'http://';

export const SERVER_URL = environment === 'production' ? `${SSL_PREFIX}${SERVER_URL_PRODUCTION}` : 'http://localhost:9000/api';

export const HANDLE_REFRESH_INTERVAL = 600000;
export const SHOWABLE_ERROR_STATUS_CODE = 666;
export const AUTHENTICATION_ERROR_STATUS_CODE = 401;
export const DOC_URL = 'https://docs.memphis.dev/memphis/memphis/overview';
export const PRIVACY_URL = 'https://app.gitbook.com/o/-MSyW3CRw3knM-KGk6G6/s/t7NJvDh5VSGZnmEsyR9h/memphis/privacy';
export const SOCKET_URL = environment === 'production' ? `ws://${SERVER_URL_PRODUCTION}:8080` : 'ws://localhost:8080';

export const GOOGLE_CLIENT_ID = '916272522459-u0f4n2lh9llsielb3l5rob3dnt1fco76.apps.googleusercontent.com';
export const REDIRECT_URI = environment === 'production' ? 'https://sandbox.memphis.dev/login' : 'http://localhost:9000/login';
export const GITHUB_CLIENT_ID = environment === 'production' ? '4dc1b3238c4d7563e426' : '51b0330eb3b34bc8f641';
