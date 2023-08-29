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

import React, { useCallback, useContext, useEffect, useState } from 'react';

import { DOCKER_UPGRADE_URL, K8S_UPGRADE_URL, LATEST_RELEASE_URL, RELEASE_DOCS_URL, RELEASE_NOTES_URL } from '../../../config';
import { GithubRequest } from '../../../services/githubRequests';
import { LOCAL_STORAGE_ENV } from '../../../const/localStorageConsts';
import { ReactComponent as UpgradeBannerIcon } from '../../../assets/images/upgradeBanner.svg';
import { ReactComponent as UpdateIcon } from '../../../assets/images/uptodateIcon.svg';
import { ReactComponent as FullLogoIcon } from '../../../assets/images/fullLogo.svg';
import Button from '../../../components/button';
import { Context } from '../../../hooks/store';
import NoteItem from './components/noteItem';
import Loader from '../../../components/loader';
import { httpRequest } from '../../../services/http';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { compareVersions } from '../../../services/valueConvertor';
import { Update } from '@material-ui/icons';

function VersionUpgrade() {
    const [state, dispatch] = useContext(Context);
    const [isLoading, setIsLoading] = useState(true);
    const [features, setFeatures] = useState({});
    const [selectedfeatures, setSelectedfeatures] = useState('Added Features');
    const [latestVersion, setLatestVersion] = useState([]);
    const [versionUrl, setversionUrl] = useState('');

    useEffect(() => {
        if (state.isLatest) {
            getSystemVersion();
        } else {
            getConfigurationValue();
        }
    }, []);

    const getSystemVersion = useCallback(async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_CLUSTER_INFO);
            if (data) {
                const latest = await GithubRequest(LATEST_RELEASE_URL);
                let is_latest = compareVersions(data.version, latest[0].name.replace('v', '').replace('-beta', '').replace('-latest', '').replace('-stable', ''));
                if (is_latest) {
                    getConfigurationValue();
                }
            }
        } catch (error) {}
    }, []);

    const getConfigurationValue = async () => {
        try {
            const latest = await GithubRequest(LATEST_RELEASE_URL);
            const latestVersion = latest[0].name?.split('-')[0];
            setLatestVersion(latestVersion);
            const data = await GithubRequest(RELEASE_NOTES_URL);
            const mdFiles = data.filter((file) => file?.name.endsWith('.md') && file?.name !== 'README.md' && file?.name?.includes(latestVersion));
            if (mdFiles.length === 0) {
                console.log('No matching files found');
                setIsLoading(false);
                return;
            }
            const mdFile = mdFiles[0];
            setversionUrl(`${RELEASE_DOCS_URL}${mdFile.name.replace('.md', '')}`);
            const file = await GithubRequest(mdFile.download_url);
            const featuresHeadlines = ['Added Features', 'Enhancements', 'Fixed bugs', 'Known issues'];
            let fetchFeatures = {};

            featuresHeadlines.map((featureHeadline) => {
                let sectionRegex = /Added features(.*?)###/s;
                switch (featureHeadline) {
                    case 'Added Features':
                        sectionRegex = /Added features(.*?)###/s;
                        break;
                    case 'Enhancements':
                        sectionRegex = /Enhancements([\s\S]*?)##/s;
                        break;
                    case 'Fixed bugs':
                        sectionRegex = /Fixed bugs(.*?)##/s;
                        break;
                    case 'Known issues':
                        sectionRegex = /Known issues([\s\S]*?)(?=##|$)/s;
                        break;
                    default:
                        sectionRegex = /Added features(.*?)###/s;
                }
                const sectionMatch = file.match(sectionRegex);
                if (sectionMatch) {
                    const featuresList = sectionMatch[1]
                        .split('\n')
                        .map((feature) => {
                            const regex = /[*-]\s*(.*)/;
                            const match = feature.match(regex);
                            if (match) {
                                return match[1].trim();
                            }
                            return null;
                        })
                        .filter((feature) => !!feature);
                    if (featuresList.length !== 0) {
                        fetchFeatures[featureHeadline] = featuresList;
                    }
                }
            });
            setFeatures(fetchFeatures);
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
                        <UpdateIcon alt="updateIcon" />
                        <div className="content">
                            <p>You are up to date.</p>
                            <span>Memphis.dev version v{state.currentVersion} is the latest version available.</span>
                        </div>
                    </div>
                </>
            ) : (
                <>
                    <div className="banner-section">
                        <UpgradeBannerIcon alt="upgradeBannerIcon" width="97%" />
                        <div className="actions">
                            <div className="current-version-wrapper">
                                <version is="x3d" style={{ cursor: !state.isLatest ? 'pointer' : 'default' }}>
                                    <p className="current-version">Current Version: v{state.currentVersion}</p>
                                </version>
                            </div>
                            <div className="logo">
                                <FullLogoIcon alt="fullLogoIcon" />
                                <div className="version-wrapper">
                                    <p>{latestVersion}</p>
                                </div>
                            </div>
                            <p className="desc-vers">A new version is available to download</p>
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
                        {Object.keys(features)?.map((key) => (
                            <Button
                                key={key}
                                width="180px"
                                height="40px"
                                placeholder={key}
                                colorType={selectedfeatures === key ? 'purple' : 'black'}
                                radiusType="circle"
                                border={selectedfeatures !== key && 'gray-light'}
                                backgroundColorType={selectedfeatures !== key ? 'white' : 'purple-light'}
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

                        {!isLoading && features[selectedfeatures]?.map((feature) => <NoteItem key={feature} feature={feature} />)}
                    </div>
                </>
            )}
        </div>
    );
}

export default VersionUpgrade;
