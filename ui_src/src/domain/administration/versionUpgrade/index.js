// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useContext, useEffect, useState } from 'react';

import { ExtractAddedFeatures, GithubRequest } from '../../../services/githubRequests';
import upgradeBanner from '../../../assets/images/upgradeBanner.svg';
import uptodateIcon from '../../../assets/images/uptodateIcon.svg';
import fullLogo from '../../../assets/images/fullLogo.svg';
import Button from '../../../components/button';
import { Context } from '../../../hooks/store';
import NoteItem from './components/noteItem';

const releaseNotes = 'https://api.github.com/repos/Memphisdev/gitbook-backup/contents/release-notes/releases';
const latestRelease = 'https://api.github.com/repos/Memphisdev/memphis/releases';

function VersionUpgrade() {
    const [state, dispatch] = useContext(Context);
    const [isLoading, setIsLoading] = useState(true);
    const [features, setFeatures] = useState([]);
    const [version, setVersion] = useState([]);
    const [versionUrl, setversionUrl] = useState('');

    useEffect(() => {
        getConfigurationValue();
    }, []);

    const getConfigurationValue = async () => {
        try {
            const latest = await GithubRequest(latestRelease);
            const version = latest[0].name;
            setVersion(version);
            const data = await GithubRequest(releaseNotes);
            const mdFiles = data.filter((file) => file.name.endsWith('.md') && file.name !== 'README.md' && file.name.includes(version));
            if (mdFiles.length === 0) {
                console.log('No matching files found');
                return;
            }
            const mdFile = mdFiles[0];
            setversionUrl(`https://docs.memphis.dev/memphis/release-notes/releases/${mdFile.name.replace('.md', '')}`);
            const file = await GithubRequest(mdFile.download_url);
            const addedFeatures = ExtractAddedFeatures(file);
            setFeatures(addedFeatures);
        } catch (err) {
            setIsLoading(false);
            return;
        }
    };

    return (
        <div className="version-upgrade-container">
            {state.isLatest ? (
                <>
                    {' '}
                    <div className="uptodate-section">
                        <img src={uptodateIcon} alt="uptodateIcon" />
                        <div className="content">
                            <p>You are up to date</p>

                            <span>Memphis version {version} is currently the newest version available.</span>
                        </div>
                    </div>
                </>
            ) : (
                <>
                    <div className="banner-section">
                        <img src={upgradeBanner} width="95%" alt="upgradeBanner" />
                        <div className="actions">
                            <div className="logo">
                                <img src={fullLogo} alt="fullLogo" />
                                <div className="version-wrapper">
                                    <p>{version}</p>
                                </div>
                            </div>
                            <p className="desc-vers">A New Version is available to download</p>
                            <Button
                                width="180px"
                                height="45px"
                                placeholder="View Full Changes"
                                colorType="black"
                                radiusType="circle"
                                backgroundColorType="white"
                                fontSize="14px"
                                fontFamily="InterSemiBold"
                                onClick={() => window.open(versionUrl, '_blank')}
                            />
                        </div>
                    </div>
                    <div className="feature-list">
                        {features.map((feature, index) => (
                            <NoteItem key={index} feature={feature} />
                        ))}
                    </div>
                </>
            )}
        </div>
    );
}

export default VersionUpgrade;
