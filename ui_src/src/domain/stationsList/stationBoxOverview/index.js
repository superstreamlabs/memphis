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

import React, { useState, useEffect } from 'react';
import { MinusOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';
import Lottie from 'lottie-react';

import { convertSecondsToDate, isCloud, parsingDate } from 'services/valueConvertor';
import activeAndHealthy from 'assets/lotties/activeAndHealthy.json';
import noActiveAndUnhealthy from 'assets/lotties/noActiveAndUnhealthy.json';
import noActiveAndHealthy from 'assets/lotties/noActiveAndHealthy.json';
import activeAndUnhealthy from 'assets/lotties/activeAndUnhealthy.json';
import { ReactComponent as RedirectIcon } from 'assets/images/redirectIcon.svg';
import { ReactComponent as ReplicasIcon } from 'assets/images/replicasIcon.svg';
import { ReactComponent as TotalMsgIcon } from 'assets/images/totalMsgIcon.svg';
import { ReactComponent as PoisonMsgIcon } from 'assets/images/poisonMsgIcon.svg';
import { ReactComponent as RemoteStorageIcon } from 'assets/images/remoteStorage.svg';
import { ReactComponent as ClockIcon } from 'assets/images/TimeFill.svg';
import { ReactComponent as UserIcon } from 'assets/images/userPerson.svg';
import { ReactComponent as SchemaIcon } from 'assets/images/schemaIconActive.svg';
import { ReactComponent as StationIcon } from 'assets/images/stationsIconActive.svg';
import { ReactComponent as RetentionIcon } from 'assets/images/retentionIcon.svg';
import { ReactComponent as PartitionIcon } from 'assets/images/partitionIcon.svg';
import OverflowTip from 'components/tooltip/overflowtip';
import CheckboxComponent from 'components/checkBox';
import TagsList from 'components/tagList';
import pathDomains from 'router';

const StationBoxOverview = ({ station, handleCheckedClick, isCheck }) => {
    const [retentionValue, setRetentionValue] = useState('');
    useEffect(() => {
        switch (station?.station?.retention_type) {
            case 'message_age_sec':
                setRetentionValue(convertSecondsToDate(station?.station?.retention_value, true));
                break;
            case 'bytes':
                setRetentionValue(`${station?.station?.retention_value} bytes`);
                break;
            case 'messages':
                setRetentionValue(`${station?.station?.retention_value} messages`);
                break;
            case 'ack_based':
                setRetentionValue('Ack');
            default:
                break;
        }
    }, []);
    return (
        <div style={{ padding: '2px' }}>
            <Link to={`${pathDomains.stations}/${station?.station?.name}`}>
                <div className="station-box-container">
                    <div className="main-section">
                        <div className="left-section">
                            <div className="station-meta">
                                <div className="header">
                                    <StationIcon />
                                    <label className="data-labels attached">Station name</label>
                                </div>
                                <div className="check-box">
                                    <CheckboxComponent checked={isCheck} id={station?.station?.name} onChange={handleCheckedClick} name={station?.station?.name} />
                                    <OverflowTip className="station-name" text={station?.station?.name} maxWidth="190px">
                                        {station?.station?.name} <label className="non-native-label">{!station?.station?.is_native && '(NATS-Compatible)'}</label>
                                    </OverflowTip>
                                </div>
                            </div>
                        </div>
                        <div className="middle-section">
                            <div className="station-meta">
                                <div className="header">
                                    <SchemaIcon />
                                    <label className="data-labels attached">Enforced schema</label>
                                </div>
                                <OverflowTip
                                    className="data-info no-text-transform"
                                    text={station?.station?.schema_name === '' ? <MinusOutlined /> : station?.station?.schema_name}
                                    width={'135px'}
                                >
                                    {station?.station?.schema_name ? station?.station?.schema_name : <MinusOutlined />}
                                </OverflowTip>
                            </div>
                        </div>
                        <div className="right-section">
                            <div className="station-meta">
                                <div className="header">
                                    <RetentionIcon />
                                    <label className="data-labels retention">Retention</label>
                                </div>
                                <OverflowTip className="data-info retention-info " text={retentionValue} width={'90px'}>
                                    {retentionValue}
                                </OverflowTip>
                            </div>
                            <div className="station-meta">
                                <div className="header">
                                    <RemoteStorageIcon />
                                    <label className="data-labels storage">Storage type</label>
                                </div>

                                <p className="data-info">{station?.station?.storage_type}</p>
                            </div>
                            {!isCloud() && (
                                <div className="station-meta">
                                    <div className="header">
                                        <ReplicasIcon />
                                        <label className="data-labels replicas">Replicas</label>
                                    </div>
                                    <p className="data-info">{station?.station?.replicas}</p>
                                </div>
                            )}
                            <div className="station-meta">
                                <div className="header">
                                    <TotalMsgIcon />
                                    <label className="data-labels total">Total messages</label>
                                </div>

                                <p className="data-info">
                                    {station.total_messages === 0 ? <MinusOutlined style={{ color: '#2E2C34' }} /> : station?.total_messages?.toLocaleString()}
                                </p>
                            </div>
                            <div className="station-meta">
                                <div className="header">
                                    <PoisonMsgIcon />
                                    <label className="data-labels total">Dead-letter messages</label>
                                </div>

                                <p className="data-info">
                                    {station.posion_messages === 0 ? <MinusOutlined style={{ color: '#2E2C34' }} /> : station?.posion_messages?.toLocaleString()}
                                </p>
                            </div>
                            <div className="station-meta">
                                <div className="header">
                                    <PartitionIcon />
                                    <label className="data-labels total">Partitions</label>
                                </div>

                                <p className="data-info">
                                    {!station?.station?.partitions_list || station?.station?.partitions_list?.length === 0
                                        ? 1
                                        : station?.station?.partitions_list?.length?.toLocaleString()}
                                </p>
                            </div>
                            <div className="station-meta poison">
                                <div className="header">
                                    <PoisonMsgIcon />
                                    <label className="data-labels">Status</label>
                                </div>
                                <div className="health-icon">
                                    {station?.has_dls_messages ? (
                                        station?.activity ? (
                                            <Lottie animationData={activeAndUnhealthy} loop={true} />
                                        ) : (
                                            <Lottie animationData={noActiveAndUnhealthy} loop={true} />
                                        )
                                    ) : station?.activity ? (
                                        <Lottie animationData={activeAndHealthy} loop={true} />
                                    ) : (
                                        <Lottie animationData={noActiveAndHealthy} loop={true} />
                                    )}
                                </div>
                            </div>
                            <div className="station-actions">
                                <div className="action">
                                    <RedirectIcon />
                                </div>
                            </div>
                        </div>
                    </div>
                    <div className="bottom-section">
                        <div className="meta-container">
                            <ClockIcon />
                            <label className="data-labels date">Created at: {parsingDate(station?.station?.created_at)}</label>
                        </div>
                        <div className="meta-container">
                            <UserIcon />
                            <label className="data-labels date">Created by: {station?.station?.created_by_username}</label>
                        </div>

                        <div className="tags-list">
                            {station?.tags.length === 0 ? (
                                <p className="data-info">
                                    <MinusOutlined />
                                </p>
                            ) : (
                                <TagsList tagsToShow={3} tags={station?.tags} />
                            )}
                        </div>
                    </div>
                </div>
            </Link>
        </div>
    );
};

export default StationBoxOverview;
