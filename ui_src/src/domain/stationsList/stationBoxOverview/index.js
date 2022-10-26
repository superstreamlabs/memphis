// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useState, useEffect } from 'react';

import MoreVertIcon from '@material-ui/icons/MoreVert';
import { MinusOutlined } from '@ant-design/icons';
import pathDomains from '../../../router';

import { convertSecondsToDate, numberWithCommas } from '../../../services/valueConvertor';
import Modal from '../../../components/modal';
import { parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';
import retentionIcon from '../../../assets/images/retentionIcon.svg';
import deleteIcon from '../../../assets/images/deleteIcon.svg';
import redirectIcon from '../../../assets/images/redirectIcon.svg';
import storageIcon from '../../../assets/images/strIcon.svg';
import replicasIcon from '../../../assets/images/replicasIcon.svg';
import totalMsgIcon from '../../../assets/images/totalMsgIcon.svg';
import poisonMsgIcon from '../../../assets/images/poisonMsgIcon.svg';
import { Link } from 'react-router-dom';
import TagsList from '../../../components/tagList';
import CheckboxComponent from '../../../components/checkBox';

const StationBoxOverview = ({ station, handleCheckedClick, removeStation, isCheck }) => {
    const [retentionValue, setRetentionValue] = useState('');

    useEffect(() => {
        switch (station.station.retention_type) {
            case 'message_age_sec':
                convertSecondsToDate(station.station.retention_value);
                setRetentionValue(convertSecondsToDate(station.station.retention_value));
                break;
            case 'bytes':
                setRetentionValue(`${station.station.retention_value} bytes`);
                break;
            case 'messages':
                setRetentionValue(`${station.station.retention_value} messages`);
                break;
            default:
                break;
        }
    }, []);

    return (
        <div>
            <div className="station-box-container">
                <div className="left-section">
                    <div className="check-box">
                        <CheckboxComponent checked={isCheck} id={station?.station?.name} onChange={handleCheckedClick} name={station?.station?.name} />
                        <p className="station-name">{station?.station?.name}</p>
                    </div>
                    <label className="data-labels">Created at {parsingDate(station.station.creation_date)}</label>
                </div>
                <div className="middle-section">
                    <div className="station-created">
                        <label className="data-labels">Created by</label>
                        <OverflowTip className="data-info" text={station.station.created_by_user} width={'100px'}>
                            {station.station.created_by_user}
                        </OverflowTip>
                    </div>
                    <div className="station-created">
                        <label className="data-labels">Tags</label>

                        <div className="tags-list">
                            {props.station.tags.length === 0 ? (
                                <p className="data-info">
                                    <MinusOutlined />
                                </p>
                            ) : (
                                <TagsList tagsToShow={4} tags={props.station.tags} />
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
                        <p className="data-info">{station.station.storage_type}</p>
                    </div>
                    <div className="station-meta">
                        <img src={replicasIcon} alt="replicas" />
                        <label className="data-labels replicas">Replicas</label>
                        <p className="data-info">{station.station.replicas}</p>
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
                        {/* <div className="action" onClick={() => modalFlip(true)}>
                                <img src={deleteIcon} />
                            </div> */}
                        <Link to={`${pathDomains.stations}/${station.station.name}`} className="action">
                            <img src={redirectIcon} />
                        </Link>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default StationBoxOverview;
