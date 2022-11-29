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

import './style.scss';

import React, { useState, useEffect } from 'react';
import { MinusOutlined } from '@ant-design/icons';
import { Link } from 'react-router-dom';

import { convertSecondsToDate, numberWithCommas } from '../../../services/valueConvertor';
import { parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';
import retentionIcon from '../../../assets/images/retentionIcon.svg';
import redirectIcon from '../../../assets/images/redirectIcon.svg';
import replicasIcon from '../../../assets/images/replicasIcon.svg';
import totalMsgIcon from '../../../assets/images/totalMsgIcon.svg';
import poisonMsgIcon from '../../../assets/images/poisonMsgIcon.svg';
import CheckboxComponent from '../../../components/checkBox';
import storageIcon from '../../../assets/images/strIcon.svg';
import TagsList from '../../../components/tagList';
import pathDomains from '../../../router';

const StationBoxOverview = ({ station, handleCheckedClick, isCheck }) => {
    const [retentionValue, setRetentionValue] = useState('');
    useEffect(() => {
        switch (station?.station?.retention_type) {
            case 'message_age_sec':
                convertSecondsToDate(station?.station?.retention_value);
                setRetentionValue(convertSecondsToDate(station?.station?.retention_value));
                break;
            case 'bytes':
                setRetentionValue(`${station?.station?.retention_value} bytes`);
                break;
            case 'messages':
                setRetentionValue(`${station?.station?.retention_value} messages`);
                break;
            default:
                break;
        }
    }, []);

    return (
        <div>
            <CheckboxComponent className="check-box-station" checked={isCheck} id={station?.station?.name} onChange={handleCheckedClick} name={station?.station?.name} />
            <Link to={`${pathDomains.stations}/${station?.station?.name}`}>
                <div className="station-box-container">
                    <div className="left-section">
                        <div className="check-box">
                            <p className="station-name">{station?.station?.name}</p>
                        </div>
                        <label className="data-labels">Created at {parsingDate(station?.station?.creation_date)}</label>
                    </div>
                    <div className="middle-section">
                        <div className="station-created">
                            <label className="data-labels">Attached Schema</label>
                            <OverflowTip
                                className="data-info"
                                text={station?.station?.schema?.name === '' ? <MinusOutlined /> : station?.station?.schema?.name}
                                width={'90px'}
                            >
                                {station?.station?.schema?.name === '' ? <MinusOutlined /> : station?.station?.schema?.name}
                            </OverflowTip>
                        </div>
                        <div className="station-created">
                            <label className="data-labels">Tags</label>

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
                    <div className="right-section">
                        <div className="station-meta">
                            <img src={retentionIcon} alt="retention" />
                            <label className="data-labels retention">Retention</label>
                            <OverflowTip className="data-info" text={retentionValue} width={'90px'}>
                                {retentionValue}
                            </OverflowTip>
                        </div>
                        <div className="station-meta">
                            <img src={storageIcon} alt="storage" />
                            <label className="data-labels storage">Storage Type</label>
                            <p className="data-info">{station?.station?.storage_type}</p>
                        </div>
                        <div className="station-meta">
                            <img src={replicasIcon} alt="replicas" />
                            <label className="data-labels replicas">Replicas</label>
                            <p className="data-info">{station?.station?.replicas}</p>
                        </div>
                        <div className="station-meta">
                            <img src={totalMsgIcon} alt="total messages" />
                            <label className="data-labels total">Total messages</label>
                            <p className="data-info">
                                {station.total_messages === 0 ? <MinusOutlined style={{ color: '#2E2C34' }} /> : numberWithCommas(station?.total_messages)}
                            </p>
                        </div>
                        <div className="station-meta">
                            <img src={poisonMsgIcon} alt="poison messages" />
                            <label className="data-labels poison">Poison messages</label>
                            <p className="data-info">{station?.posion_messages === 0 ? <MinusOutlined /> : numberWithCommas(station?.posion_messages)}</p>
                        </div>
                        <div className="station-actions">
                            <div className="action">
                                <img src={redirectIcon} alt="redirectIcon" />
                            </div>
                        </div>
                    </div>
                </div>
            </Link>
        </div>
    );
};

export default StationBoxOverview;
