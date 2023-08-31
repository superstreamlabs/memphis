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

import React, { useContext } from 'react';
import { useHistory } from 'react-router-dom';
import { KeyboardArrowRightRounded } from '@material-ui/icons';
import Lottie from 'lottie-react';

import noActiveAndUnhealthy from '../../../assets/lotties/noActiveAndUnhealthy.json';
import { isCloud, parsingDate } from '../../../services/valueConvertor';
import noActiveAndHealthy from '../../../assets/lotties/noActiveAndHealthy.json';
import activeAndUnhealthy from '../../../assets/lotties/activeAndUnhealthy.json';
import activeAndHealthy from '../../../assets/lotties/activeAndHealthy.json';
import OverflowTip from '../../../components/tooltip/overflowtip';
import NoStations from '../../../assets/images/noStations.svg';
import Button from '../../../components/button';
import Filter from '../../../components/filter';
import { Context } from '../../../hooks/store';
import pathDomains from '../../../router';
import { Virtuoso } from 'react-virtuoso';

const Stations = ({ createStationTrigger }) => {
    const [state, dispatch] = useContext(Context);
    const history = useHistory();

    const goToStation = (stationName) => {
        history.push(`${pathDomains.stations}/${stationName}`);
    };
    return (
        <div className="overview-components-wrapper">
            <div className="stations-container">
                <div className="overview-components-header stations-header">
                    <p> Stations {state?.monitor_data?.stations?.length > 0 && `(${state?.monitor_data?.stations?.length})`}</p>
                    <label>A station is a distributed unit that stores messages</label>
                </div>
                <div className="err-stations-list">
                    {state?.monitor_data?.stations?.length > 0 ? (
                        <>
                            <div className={!isCloud() ? 'coulmns-table' : 'coulmns-table coulmns-table-cloud'}>
                                <span className="station-name">Name</span>
                                {!isCloud() && <span>Creation date</span>}
                                <span className="title-center">Mssages</span>
                                <span className="title-center">Partitions</span>
                                <span className="title-center">Status</span>
                                <span></span>
                            </div>
                            <div className={!isCloud() ? 'rows-wrapper' : 'rows-wrapper rows-wrapper-cloud'}>
                                <Virtuoso
                                    data={state?.monitor_data?.stations}
                                    overscan={100}
                                    itemContent={(index, station) => (
                                        <div className={index % 2 === 0 ? 'stations-row' : 'stations-row even'} key={index} onClick={() => goToStation(station.name)}>
                                            <OverflowTip className="station-details station-name" text={station.name}>
                                                {station.name}
                                                {isCloud() && <span className="creates">{parsingDate(station.created_at)}</span>}
                                            </OverflowTip>
                                            {!isCloud() && (
                                                <OverflowTip className="station-creation" text={parsingDate(station.created_at)}>
                                                    {parsingDate(station.created_at)}
                                                </OverflowTip>
                                            )}
                                            <OverflowTip className="station-details total" text={station.total_messages?.toLocaleString()}>
                                                <span className="centered">{station.total_messages?.toLocaleString()}</span>
                                            </OverflowTip>
                                            <div className="station-details total">
                                                <span className="centered">
                                                    {!station?.partitions_list || station?.partitions_list?.length === 0
                                                        ? 1
                                                        : station?.partitions_list?.length?.toLocaleString()}
                                                </span>
                                            </div>
                                            <div className={!isCloud() ? 'centered lottie' : 'centered lottie-cloud'}>
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
                                            <div className="centered">
                                                <div className="staion-link">
                                                    <span>View</span>
                                                    <KeyboardArrowRightRounded />
                                                </div>
                                            </div>
                                        </div>
                                    )}
                                />
                            </div>
                        </>
                    ) : (
                        <div className="empty-stations-container">
                            <img src={NoStations} alt="no stations" onClick={() => createStationTrigger(true)} />
                            <p>No stations yet</p>
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
                                onClick={() => createStationTrigger(true)}
                            />
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

export default Stations;
