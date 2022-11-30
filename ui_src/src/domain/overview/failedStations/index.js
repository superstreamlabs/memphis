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

import React, { useContext, useRef, useState } from 'react';
import { Link, useHistory } from 'react-router-dom';
import { KeyboardArrowRightRounded } from '@material-ui/icons';

import { numberWithCommas, parsingDate } from '../../../services/valueConvertor';
import OverflowTip from '../../../components/tooltip/overflowtip';
import Modal from '../../../components/modal';
import Button from '../../../components/button';
import CreateStationForm from '../../../components/createStationForm';
import stationImg from '../../../assets/images/stationsIconActive.svg';
import staionLink from '../../../assets/images/staionLink.svg';
import NoStations from '../../../assets/images/noStations.svg';
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';

const FailedStations = () => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();
    const createStationRef = useRef(null);
    const [open, modalFlip] = useState(false);
    const [creatingProsessd, setCreatingProsessd] = useState(false);

    const goToStation = (stationName) => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };

    return (
        <div className="overview-wrapper failed-stations-container">
            <p className="overview-components-header" id="e2e-overview-station-list">
                Stations
            </p>
            <div className="err-stations-list">
                {state?.monitor_data?.stations?.length > 0 ? (
                    <div className="coulmns-table">
                        <span style={{ width: '100px' }}>Name</span>
                        <span style={{ width: '200px' }}>Creation date</span>
                        <span style={{ width: '120px' }}>Total messages</span>
                        <span style={{ width: '120px' }}>Poison messages</span>
                        <span style={{ width: '120px' }}></span>
                    </div>
                ) : (
                    <div className="empty-stations-container">
                        <img src={NoStations} alt="no stations" />
                        <div>
                            <p>No Stations Found</p>
                            <Button
                                className="modal-btn"
                                width="160px"
                                height="34px"
                                placeholder={'Create new station'}
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="12px"
                                fontWeight="600"
                                aria-haspopup="true"
                                onClick={() => modalFlip(true)}
                            />
                        </div>
                        <Modal
                            header={
                                <div className="modal-header">
                                    <div className="header-img-container">
                                        <img className="headerImage" src={stationImg} alt="stationImg" />
                                    </div>
                                    <p>Create new station</p>
                                    <label>A station is a distributed unit that stores the produced data.</label>
                                </div>
                            }
                            height="540px"
                            width="560px"
                            rBtnText="Add"
                            lBtnText="Cancel"
                            lBtnClick={() => {
                                modalFlip(false);
                            }}
                            rBtnClick={() => {
                                createStationRef.current();
                            }}
                            clickOutside={() => modalFlip(false)}
                            open={open}
                            isLoading={creatingProsessd}
                        >
                            <CreateStationForm createStationFormRef={createStationRef} handleClick={(e) => setCreatingProsessd(e)} />
                        </Modal>{' '}
                    </div>
                )}
                <div className="rows-wrapper">
                    {state?.monitor_data?.stations?.map((station, index) => {
                        return (
                            <div className="stations-row" key={index} onClick={() => goToStation(station.name)}>
                                <OverflowTip className="station-details" text={station.name} width={'100px'}>
                                    {station.name}
                                </OverflowTip>
                                <OverflowTip className="station-creation" text={parsingDate(station.creation_date)} width={'200px'}>
                                    {parsingDate(station.creation_date)}
                                </OverflowTip>
                                <span className="station-details centered">{numberWithCommas(station.total_messages)}</span>
                                <span className="station-details centered">{numberWithCommas(station.posion_messages)}</span>
                                <div className="link-wrapper">
                                    <div className="staion-link">
                                        <span>View Station</span>
                                        <KeyboardArrowRightRounded />
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

export default FailedStations;
