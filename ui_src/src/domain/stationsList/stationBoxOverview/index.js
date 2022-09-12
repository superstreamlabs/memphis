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
import DeleteOutline from '@material-ui/icons/DeleteOutline';
import MenuItem from '@material-ui/core/MenuItem';
import Popover from '@material-ui/core/Popover';
import { MinusOutlined } from '@ant-design/icons';
import pathDomains from '../../../router';

import { convertSecondsToDate } from '../../../services/valueConvertor';
import Modal from '../../../components/modal';
import { parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';
import retentionImg from '../../../assets/images/retention.svg';
import storageImg from '../../../assets/images/storage.svg';
import replicasImg from '../../../assets/images/replicas.svg';
import totalMsgImg from '../../../assets/images/total_msg.svg';
import poisonMsgImg from '../../../assets/images/poison_msg.svg';
import { Link } from 'react-router-dom';
import stationsIcon from '../../../assets/images/stationsIcon.svg';

const StationBoxOverview = (props) => {
    const [modalIsOpen, modalFlip] = useState(false);
    const [anchorEl, setAnchorEl] = useState(null);
    const open = Boolean(anchorEl);
    const [retentionValue, setRetentionValue] = useState('');

    useEffect(() => {
        switch (props.station.station.retention_type) {
            case 'message_age_sec':
                convertSecondsToDate(props.station.station.retention_value);
                setRetentionValue(convertSecondsToDate(props.station.station.retention_value));
                break;
            case 'bytes':
                setRetentionValue(`${props.station.station.retention_value} bytes`);
                break;
            case 'messages':
                setRetentionValue(`${props.station.station.retention_value} messages`);
                break;
            default:
                break;
        }
    }, []);

    const handleClickMenu = (event) => {
        setAnchorEl(event.currentTarget);
    };

    const handleCloseMenu = () => {
        setAnchorEl(null);
    };

    return (
        <div>
            <Link className="station-box-container" to={`${pathDomains.stations}/${props.station.station.name}`}>
                <div className="left-section">
                    <p className="station-name">{props.station?.station?.name}</p>
                    <label className="data-labels">Created: {parsingDate(props.station.station.creation_date)}</label>
                </div>
                <div className="middle-section">
                    <div className="station-created">
                        <label className="data-labels">Created by</label>
                        <OverflowTip text={props.station.station.created_by_user} width={'100px'}>
                            <p className="data-info">{props.station.station.created_by_user}</p>
                        </OverflowTip>
                    </div>
                </div>
                <div className="right-section">
                    <div className="station-meta">
                        <img src={retentionImg} alt="retention" />
                        <label className="data-labels">Retention</label>
                        <p className="data-info">
                            <OverflowTip text={retentionValue} width={'90px'}>
                                {retentionValue}
                            </OverflowTip>
                        </p>
                    </div>
                    <div className="station-meta">
                        <img src={storageImg} alt="storage" />
                        <label className="data-labels">Storage Type</label>
                        <p className="data-info">{props.station.station.storage_type}</p>
                    </div>
                    <div className="station-meta">
                        <img src={replicasImg} alt="replicas" />
                        <label className="data-labels">Replicas</label>
                        <p className="data-info">{props.station.station.replicas}</p>
                    </div>
                    <div className="station-meta">
                        <img src={totalMsgImg} alt="total messages" />
                        <label className="data-labels">Total messages</label>
                        <p className="data-info">
                            {props.station.total_messages === 0 ? <MinusOutlined style={{ color: '#2E2C34' }} /> : props?.station?.total_messages}
                        </p>
                    </div>
                    <div className="station-meta">
                        <img src={poisonMsgImg} alt="poison messages" />
                        <label className="data-labels">Poison messages</label>
                        <p className="data-info">{props?.station?.posion_messages === 0 ? <MinusOutlined /> : props?.station?.posion_messages}</p>
                    </div>
                    <MoreVertIcon
                        aria-controls="long-button"
                        aria-haspopup="true"
                        onClick={(e) => {
                            e.preventDefault();
                            handleClickMenu(e);
                        }}
                        className="threedots-menu"
                    />
                </div>
            </Link>

            <Popover id="long-menu" classes={{ paper: 'Menu c' }} anchorEl={anchorEl} onClose={handleCloseMenu} open={open}>
                <MenuItem
                    onClick={() => {
                        modalFlip(true);
                        handleCloseMenu();
                    }}
                >
                    <DeleteOutline className="menu-item-icon" />
                    <label id="e2e-tests-remove-stations" className="menu-item-label">
                        Remove
                    </label>
                </MenuItem>
                <MenuItem>
                    <img src={stationsIcon} alt="stationsIcon" style={{ height: '15px', width: '15px' }} />
                    <Link
                        id="e2e-tests-remove-stations"
                        className="menu-item-label"
                        style={{ color: 'black' }}
                        to={`${pathDomains.stations}/${props.station.station.name}`}
                    >
                        Overview
                    </Link>
                </MenuItem>
            </Popover>
            <Modal
                header="Remove station"
                height="100px"
                minWidth="460px"
                rBtnText="Cancel"
                lBtnText="Remove"
                lBtnClick={() => {
                    props.removeStation();
                    modalFlip(false);
                }}
                closeAction={() => modalFlip(false)}
                clickOutside={() => modalFlip(false)}
                rBtnClick={() => modalFlip(false)}
                open={modalIsOpen}
            >
                <label>
                    Are you sure you want to delete "<b>{props.station.station.name}</b>" station?
                </label>
                <br />
            </Modal>
        </div>
    );
};

export default StationBoxOverview;
