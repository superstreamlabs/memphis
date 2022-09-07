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
// limitations under the License.

import './style.scss';

import React, { useState } from 'react';
import { ChevronRightOutlined } from '@material-ui/icons';

import copyClipboard from '../../../../assets/images/copyClipboard.svg';
import Copy from '../../../../assets/images/copy.svg';
import Copied from '../../../../assets/images/copied.svg';
import comingSoonBox from '../../../../assets/images/comingSoonBox.svg';
import redirectIcon from '../../../../assets/images/redirectIcon.svg';
import videoIcon from '../../../../assets/images/videoIcon.svg';
import docsPurple from '../../../../assets/images/docsPurple.svg';
import { Link } from 'react-router-dom';

const InstallationCommand = ({ steps, showLinks, videoLink, docsLink }) => {
    const [copied, setCopied] = useState(null);

    const handleCopy = (key, data) => {
        setCopied(key);
        navigator.clipboard.writeText(data);
        setTimeout(() => {
            setCopied(null);
        }, 3000);
    };

    return (
        <div className="installation-command">
            {steps.length > 0 && (
                <div className="steps">
                    {steps.map((value, key) => {
                        return (
                            <div className="step-wrapper" key={key}>
                                <p className="step-title">{value.title}</p>
                                <div className="step-command">
                                    <span>{value.command}</span>
                                    {value.icon === 'copy' && <img src={copied === key ? Copied : Copy} onClick={() => handleCopy(key, value.command)} />}
                                    {value.icon === 'link' && (
                                        <Link to={{ pathname: 'http://localhost:9000' }} target="_blank">
                                            <img src={redirectIcon} />
                                        </Link>
                                    )}
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
            {steps.length === 0 && (
                <div className="coming-soon-wrapper">
                    <img src={comingSoonBox} />
                    <p>Coming soon...</p>
                </div>
            )}
            {showLinks && (
                <div className="links">
                    <Link to={{ pathname: videoLink }} target="_blank">
                        <div className="link-wrapper">
                            <img src={videoIcon} />
                            <p>Installation video</p>
                            <ChevronRightOutlined />
                        </div>
                    </Link>
                    <Link to={{ pathname: docsLink }} target="_blank">
                        <div className="link-wrapper">
                            <img width={25} height={22} src={docsPurple} />
                            <p>Link to docs</p>
                            <ChevronRightOutlined />
                        </div>
                    </Link>
                </div>
            )}
        </div>
    );
};

export default InstallationCommand;
