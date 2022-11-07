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

import React, { useState } from 'react';

import copy from '../../assets/images/copy.svg';
import copiedIcon from '../../assets/images/copied.svg';

const Copy = ({ data, key }) => {
    const [copied, setCopied] = useState(null);

    const handleCopyWithKey = (key, data) => {
        setCopied(key);
        navigator.clipboard.writeText(data);
        setTimeout(() => {
            setCopied(null);
        }, 3000);
    };
    const handleCopy = (data) => {
        setCopied(true);
        navigator.clipboard.writeText(data);
        setTimeout(() => {
            setCopied(false);
        }, 3000);
    };
    return (
        <>
            {key && <img alt="copy" style={{ cursor: 'pointer' }} src={copied === key ? copiedIcon : copy} onClick={() => handleCopyWithKey(key, data)} />}
            {!key && <img alt="copy" style={{ cursor: 'pointer' }} src={copied ? copiedIcon : copy} onClick={() => handleCopy(data)} />}
        </>
    );
};

export default Copy;
