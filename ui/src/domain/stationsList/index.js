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

import React, { useEffect, useContext, useState, useRef, useCallback, Fragment } from 'react';
import ClickAwayListener from '@material-ui/core/ClickAwayListener';
import CircularProgress from '@material-ui/core/CircularProgress';
import EditOutlined from '@material-ui/icons/EditOutlined';
import { useHistory } from 'react-router-dom';

import CreateStationDetails from '../../components/createStationDetails';
import { ApiEndpoints } from '../../const/apiEndpoints';
import StationBoxOverview from './stationBoxOverview';
import emptyList from '../../assets/images/emptyList.svg';
import { httpRequest } from '../../services/http';
import Button from '../../components/button';
import { Context } from '../../hooks/store';
import Modal from '../../components/modal';
import pathDomains from '../../router';
import Loader from '../../components/loader';
import { SOCKET_URL } from '../../config';
import { LOCAL_STORAGE_TOKEN } from '../../const/localStorageConsts';
import { parsingDate } from '../../services/valueConvertor';

const StationsList = () => {
    const url = window.location.href;
    const urlfactoryName = url.split('factories/')[1].split('/')[0];
    const history = useHistory();
    const botId = 1;

    const [state, dispatch] = useContext(Context);
    const [editName, seteditName] = useState(false);
    const [editDescription, seteditDescription] = useState(false);
    const [modalIsOpen, modalFlip] = useState(false);
    const [factoryDetails, setFactoryDetails] = useState();
    const [factoryName, setFactoryName] = useState('');
    const [factoryDescription, setFactoryDescription] = useState('');
    const [isLoading, setisLoading] = useState(false);
    const createStationRef = useRef(null);
    const [parseDate, setParseDate] = useState('');
    const [botUrl, SetBotUrl] = useState('');

    const getFactoryDetails = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', `${ApiEndpoints.GEL_FACTORY}?factory_name=${urlfactoryName}`);
            setBotImage(data.user_avatar_id || botId);
            setParseDate(parsingDate(data.creation_date));
            setFactoryDetails(data);
            setFactoryName(data.name);
            setFactoryDescription(data.description);
            setisLoading(false);
        } catch (error) {
            setisLoading(false);
            if (error.status === 404) {
                history.push(pathDomains.factoriesList);
            }
        }
    };

    const setBotImage = (botId) => {
        SetBotUrl(require(`../../assets/images/bots/${botId}.svg`));
    };

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'factories' });
        getFactoryDetails();
    }, []);

    const handleRegisterToFactory = useCallback(
        (factoryName) => {
            state.socket?.emit('register_factory_overview_data', factoryName);
        },
        [state.socket]
    );

    useEffect(() => {
        state.socket?.on(`factory_overview_data_${urlfactoryName}`, (data) => {
            setBotImage(data.user_avatar_id || botId);
            setParseDate(parsingDate(data.creation_date));
            setFactoryDetails(data);
        });

        state.socket?.on('error', (error) => {
            history.push(pathDomains.factoriesList);
        });

        setTimeout(() => {
            handleRegisterToFactory(urlfactoryName);
        }, 1000);

        return () => {
            state.socket?.emit('deregister');
        };
    }, [state.socket]);

    const handleEditName = useCallback(() => {
        state.socket?.emit('deregister');
        seteditName(true);
    }, [state.socket]);

    const handleEditDescription = useCallback(() => {
        state.socket?.emit('deregister');
        seteditDescription(true);
    }, [state.socket]);

    const handleEditNameBlur = async (e) => {
        if (!e.target.value || e.target.value === factoryDetails.name || e.target.value === '') {
            setFactoryName(factoryDetails.name);
            handleRegisterToFactory(factoryDetails.name);
            seteditName(false);
        } else {
            try {
                await httpRequest('PUT', ApiEndpoints.EDIT_FACTORY, {
                    factory_name: factoryDetails.name,
                    factory_new_name: e.target.value
                });
                handleRegisterToFactory(e.target.value);
                setFactoryDetails({ ...factoryDetails, name: e.target.value });
                seteditName(false);
                history.push(`${pathDomains.factoriesList}/${e.target.value}`);
            } catch (err) {
                setFactoryName(factoryDetails.name);
            }
        }
    };

    const handleEditNameChange = (e) => {
        setFactoryName(e.target.value);
    };

    const handleEditDescriptionBlur = async (e) => {
        if (e.target.value === factoryDetails.description) {
            handleRegisterToFactory(factoryName);
            seteditDescription(false);
        } else {
            try {
                await httpRequest('PUT', ApiEndpoints.EDIT_FACTORY, {
                    factory_name: factoryDetails.name,
                    factory_new_description: e.target.value
                });
                setFactoryDetails({ ...factoryDetails, description: e.target.value });
                seteditDescription(false);
            } catch (err) {
                setFactoryDescription(factoryDetails.description);
            }
        }
    };

    const handleEditDescriptionChange = (e) => {
        setFactoryDescription(e.target.value);
    };

    const removeStation = async (stationName) => {
        const updatedStationList = factoryDetails?.stations.filter((item) => item.name !== stationName);
        setFactoryDetails({ ...factoryDetails, stations: updatedStationList });
    };

    return (
        <div className="factory-details-container">
            {isLoading && (
                <div className="loader-uploading">
                    <Loader />
                </div>
            )}
            {!isLoading && (
                <Fragment>
                    <div className="factory-details-header">
                        <div className="left-side">
                            {!editName && (
                                <h1 className="main-header-h1">
                                    {factoryName || 'Insert Factory name'}
                                    <span id="e2e-tests-edit-name" className="edit-icon" onClick={() => handleEditName()}>
                                        <EditOutlined />
                                    </span>
                                </h1>
                            )}
                            {editName && (
                                <ClickAwayListener onClickAway={handleEditNameBlur}>
                                    <div className="edit-input-name">
                                        <input onBlur={handleEditNameBlur} onChange={handleEditNameChange} value={factoryName} />
                                    </div>
                                </ClickAwayListener>
                            )}
                            {!editDescription && (
                                <div className="description">
                                    {<p>{factoryDescription || 'Insert your description...'}</p>}
                                    <span id="e2e-tests-edit-description" className="edit-icon" onClick={() => handleEditDescription()}>
                                        <EditOutlined />
                                    </span>
                                </div>
                            )}
                            {editDescription && (
                                <ClickAwayListener onClickAway={handleEditDescriptionBlur}>
                                    <div id="e2e-tests-insert-description">
                                        <textarea onBlur={handleEditDescriptionBlur} onChange={handleEditDescriptionChange} value={factoryDescription} />
                                    </div>
                                </ClickAwayListener>
                            )}

                            <div className="factory-owner">
                                <div className="user-avatar">
                                    <img src={botUrl} width={25} height={25} alt="bot"></img>
                                </div>
                                <div className="user-details">
                                    <p>{factoryDetails?.created_by_user}</p>
                                    <span>{parseDate}</span>
                                </div>
                            </div>

                            <div className="factories-length">
                                <h1>Stations ({factoryDetails?.stations?.length || 0})</h1>
                            </div>
                        </div>
                        <div className="right-side">
                            <Button
                                className="modal-btn"
                                width="150px"
                                height="36px"
                                placeholder="Create a station"
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="14px"
                                fontWeight="bold"
                                aria-controls="usecse-menu"
                                aria-haspopup="true"
                                onClick={() => modalFlip(true)}
                            />
                        </div>
                    </div>
                    <div className="stations-content">
                        {factoryDetails?.stations?.length > 0 &&
                            factoryDetails?.stations?.map((station, key) => (
                                <StationBoxOverview key={station.id} station={station} removeStation={() => removeStation(station.name)} />
                            ))}
                        {!isLoading && factoryDetails?.stations.length === 0 && (
                            <div className="no-station-to-display">
                                <img src={emptyList} width="100" height="100" alt="emptyList" />
                                <p>There are no stations yet</p>
                                <p className="sub-title">Get started by creating a station</p>
                                <Button
                                    className="modal-btn"
                                    width="240px"
                                    height="50px"
                                    placeholder="Create your first station"
                                    colorType="white"
                                    radiusType="circle"
                                    backgroundColorType="purple"
                                    fontSize="12px"
                                    fontWeight="600"
                                    aria-controls="usecse-menu"
                                    aria-haspopup="true"
                                    onClick={() => modalFlip(true)}
                                />
                            </div>
                        )}
                    </div>
                    <Modal
                        header="Your station details"
                        rBtnText="Add"
                        lBtnText="Cancel"
                        lBtnClick={() => {
                            modalFlip(false);
                        }}
                        clickOutside={() => modalFlip(false)}
                        rBtnClick={() => {
                            createStationRef.current();
                        }}
                        open={modalIsOpen}
                    >
                        <CreateStationDetails createStationRef={createStationRef} factoryName={factoryName} />
                    </Modal>
                </Fragment>
            )}
        </div>
    );
};

export default StationsList;
