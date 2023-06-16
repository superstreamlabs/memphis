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

import { DOCKER_UPGRADE_URL, K8S_UPGRADE_URL, LATEST_RELEASE_URL, RELEASE_DOCS_URL, RELEASE_NOTES_URL } from '../../../config';
import { ExtractFeatures, GithubRequest } from '../../../services/githubRequests';
import { LOCAL_STORAGE_ENV } from '../../../const/localStorageConsts';
import upgradeBanner from '../../../assets/images/upgradeBanner.svg';
import uptodateIcon from '../../../assets/images/uptodateIcon.svg';
import fullLogo from '../../../assets/images/fullLogo.svg';
import Button from '../../../components/button';
import { Context } from '../../../hooks/store';
import NoteItem from './components/noteItem';
import Loader from '../../../components/loader';

function VersionUpgrade() {
    const [state] = useContext(Context);
    const [isLoading, setIsLoading] = useState(true);
    const [features, setFeatures] = useState({});
    const [selectedfeatures, setSelectedfeatures] = useState('Added Features');
    const [version, setVersion] = useState([]);
    const [versionUrl, setversionUrl] = useState('');

    useEffect(() => {
        getConfigurationValue();
    }, []);

    const getConfigurationValue = async () => {
        try {
            const latest = await GithubRequest(LATEST_RELEASE_URL);
            const version = latest[0].name?.split('-')[0];
            setVersion(version);
            const data = await GithubRequest(RELEASE_NOTES_URL);
            const mdFiles = data.filter((file) => file?.name.endsWith('.md') && file?.name !== 'README.md' && file?.name?.includes(version));
            if (mdFiles.length === 0) {
                console.log('No matching files found');
                setIsLoading(false);
                return;
            }
            const mdFile = mdFiles[0];
            setversionUrl(`${RELEASE_DOCS_URL}${mdFile.name.replace('.md', '')}`);
            const file = await GithubRequest(mdFile.download_url);
            const featuresHeadlines = ['Added Features', 'Enhancements', 'Fixed Bugs', 'Known Issues'];
            let regex = /###\s*!\[:sparkles:\].*?Added\s*features\s*(.*?)\s*(?=###|$)/is;
            let FetchFeatures = {};

            featuresHeadlines.map((featureHeadline) => {
                switch (featureHeadline) {
                    case 'Added Features':
                        regex = /###\s*!\[:sparkles:\].*?Added\s*features\s*(.*?)\s*(?=###|$)/is;
                        break;
                    case 'Enhancements':
                        regex = /###\s*!\[:sparkles:\].*?Enhancements\s*(.*?)\s*(?=###|$)/is;
                        break;
                    case 'Fixed Bugs':
                        regex = /###\s*!\[:sparkles:\].*?Fixed\s*bugs\s*(.*?)\s*(?=###|$)/is;
                        break;
                    case 'Known Issues':
                        regex = /###\s*!\[:sparkles:\].*?Known\s*issues\s*(.*?)\s*(?=###|$)/is;
                        break;
                    default:
                        break;
                }

                const featuresList = ExtractFeatures(file, regex);
                if (featuresList.length !== 0) {
                    FetchFeatures[featureHeadline] = featuresList;
                }
            });
            setFeatures(FetchFeatures);
            setIsLoading(false);
        } catch (err) {
            setIsLoading(false);
            return;
        }
    };

    const howToUpgrade = () => {
        localStorage.getItem(LOCAL_STORAGE_ENV) === 'docker' ? window.open(DOCKER_UPGRADE_URL, '_blank') : window.open(K8S_UPGRADE_URL, '_blank');
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
                        <img src={upgradeBanner} width="97%" alt="upgradeBanner" />
                        <div className="actions">
                            <div className="current-version-wrapper">
                                <version is="x3d" style={{ cursor: !state.isLatest ? 'pointer' : 'default' }}>
                                    <p className="current-version">Current Version: v{state.currentVersion}</p>
                                </version>
                            </div>
                            <div className="logo">
                                <img src={fullLogo} alt="fullLogo" />
                                <div className="version-wrapper">
                                    <p>{version}</p>
                                </div>
                            </div>
                            <p className="desc-vers">A New Version is available to download</p>
                            <div className="buttons">
                                <Button
                                    width="180px"
                                    height="40px"
                                    placeholder="View Full Changes"
                                    colorType="black"
                                    radiusType="circle"
                                    backgroundColorType="white"
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    onClick={() => window.open(versionUrl, '_blank')}
                                />
                                <Button
                                    width="180px"
                                    height="40px"
                                    placeholder="How to upgrade"
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="12px"
                                    fontFamily="InterSemiBold"
                                    onClick={() => howToUpgrade()}
                                />
                            </div>
                        </div>
                    </div>
                    <div className="feature-buttons">
                        {Object.keys(features).map((key) => (
                            <Button
                                width="180px"
                                height="40px"
                                placeholder={key}
                                colorType={selectedfeatures === key ? 'purple' : 'black'}
                                radiusType="circle"
                                border={selectedfeatures !== key && 'black'}
                                backgroundColorType={selectedfeatures !== key && 'white'}
                                fontSize="12px"
                                fontFamily="InterSemiBold"
                                onClick={() => setSelectedfeatures(key)}
                            />
                        ))}
                    </div>
                    <div className="feature-list">
                        {isLoading && (
                            <div className="loading">
                                <Loader background={false} />
                            </div>
                        )}

                        {!isLoading && features[selectedfeatures].map((feature) => <NoteItem key={feature} feature={feature} />)}
                    </div>
                </>
            )}
        </div>
    );
}

export default VersionUpgrade;
