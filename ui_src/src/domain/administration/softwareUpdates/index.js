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
import Button from '../../../components/button';
import CloudModal from '../../../components/cloudModal';
import { Context } from '../../../hooks/store';
import { ReactComponent as LogoTexeMemphis } from '../../../assets/images/logoTexeMemphis.svg';
import { ReactComponent as RedirectWhiteIcon } from '../../../assets/images/exportWhite.svg';
import { ReactComponent as DocumentIcon } from '../../../assets/images/documentGroupIcon.svg';
import { ReactComponent as DisordIcon } from '../../../assets/images/discordGroupIcon.svg';
import { ReactComponent as WindowIcon } from '../../../assets/images/windowGroupIcon.svg';
import { ApiEndpoints } from '../../../const/apiEndpoints';
import { httpRequest } from '../../../services/http';
import { GithubRequest } from '../../../services/githubRequests';
import { LATEST_RELEASE_URL } from '../../../config';
import { compareVersions } from '../../../services/valueConvertor';

function SoftwareUpates({}) {
    const [state, dispatch] = useContext(Context);
    const [isCloudModalOpen, setIsCloudModalOpen] = useState(false);
    const [systemData, setSystemData] = useState({});
    const [version, setVersion] = useState('v' + state?.currentVersion);
    const [isUpdateAvailable, setIsUpdateAvailable] = useState(false);
    const [latestVersionUrl, setLatestVersionUrl] = useState('');

    const systemDataComponents = [
        { title: 'Amount of brokers', value: systemData?.total_amount_brokers },
        { title: 'total stations', value: systemData?.total_stations },
        { title: 'total users', value: systemData?.total_users },
        { title: 'total schemas', value: systemData?.total_schemas }
    ];

    const informationPanelData = [
        {
            icon: <DocumentIcon />,
            title: 'Read Our documentation',
            description: (
                <span>
                    Read our documentation to learn more about <span> Memphis.dev</span>
                </span>
            ),
            onClick: () => {
                window.open('https://docs.memphis.dev/memphis/getting-started/readme', '_blank');
            }
        },
        {
            icon: <DisordIcon />,
            title: 'Join our Discord',
            description: (
                <span>
                    Find <span>Memphis.dev's</span> Open-Source contributors and maintainers here
                </span>
            ),
            onClick: () => {
                window.open('https://memphis.dev/discord', '_blank');
            }
        },
        {
            icon: <WindowIcon />,
            title: 'Open a service request',
            description: <span>If you have any questions or need assistance. </span>,
            onClick: () => {
                setIsCloudModalOpen(true);
            }
        }
    ];

    const genrateInformationPanel = (item, index) => (
        <div className="item-component" key={index} onClick={() => item?.onClick()}>
            {item?.icon}
            <p>{item?.title}</p>
            {item?.description}
        </div>
    );

    useEffect(() => {
        getSystemGeneralInfo();
        getSystemVersion();
    }, []);

    const getSystemGeneralInfo = async () => {
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GET_SYSTEM_GENERAL_INFO}`);
            setSystemData(data);
        } catch (err) {
            return;
        }
    };

    const getSystemVersion = async () => {
        try {
            const data = await httpRequest('GET', ApiEndpoints.GET_CLUSTER_INFO);
            if (data) {
                setVersion('v' + data?.version);
                const latest = await GithubRequest(LATEST_RELEASE_URL);
                setLatestVersionUrl(latest[0].html_url);
                const is_latest = compareVersions(data?.version, latest[0].name.replace('v', '').replace('-beta', '').replace('-latest', '').replace('-stable', ''));
                setIsUpdateAvailable(!is_latest);
            }
        } catch (error) {}
    };

    return (
        <div className="softwate-updates-container">
            <div className="rows">
                <div className="item-component">
                    <div className="title-component">
                        <div className="versions" onClick={() => isUpdateAvailable && window.open(latestVersionUrl, '_blank')}>
                            <LogoTexeMemphis alt="Memphis logo" width="300px" />
                            <label className="curr-version">{version}</label>
                            {isUpdateAvailable && <div className="red-dot" />}
                        </div>
                        <Button
                            width="200px"
                            height="36px"
                            placeholder={
                                <span className="change-log">
                                    <label>View Change log</label>
                                    <RedirectWhiteIcon alt="redirect" />
                                </span>
                            }
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType={'purple'}
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            onClick={() => {
                                window.open('https://docs.memphis.dev/memphis/release-notes/releases', '_blank');
                            }}
                        />
                    </div>
                </div>
                <div className="statistics">
                    {systemDataComponents.map((item, index) => {
                        return (
                            <div className="item-component wrapper" key={`${item}-${index}`}>
                                <label className="title">{item.title}</label>
                                <label className="numbers">{item.value}</label>
                            </div>
                        );
                    })}
                </div>
                <div className="charts">{informationPanelData.map((item, index) => genrateInformationPanel(item, index))}</div>
            </div>
            <CloudModal type={'bundle'} open={isCloudModalOpen} handleClose={() => setIsCloudModalOpen(false)} />
        </div>
    );
}

export default SoftwareUpates;
