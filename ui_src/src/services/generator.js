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

export const generator = () => {
    const string = 'abcdefghijklmnopqrstuvwxyz';
    const numeric = '0123456789';
    const length = 9;
    const formValid = +length > 0;
    if (!formValid) {
        return;
    }
    let character = '';
    let password = '';
    while (password.length < length) {
        const entity1 = Math.ceil(string.length * Math.random() * Math.random());
        const entity2 = Math.ceil(numeric.length * Math.random() * Math.random());
        let hold = string.charAt(entity1);
        character += hold;
        character += numeric.charAt(entity2);
        password = character;
    }
    password = password
        .split('')
        .sort(() => {
            return 0.5 - Math.random();
        })
        .join('');
    return password.substr(0, length);
};
