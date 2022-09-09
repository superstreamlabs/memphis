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

import React, { useEffect, useContext, useState, useRef } from 'react';

import CreateFactoryDetails from './createFactoryDetails';
import emptyList from '../../assets/images/emptyList.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import { httpRequest } from '../../services/http';
import Button from '../../components/button';
import Loader from '../../components/loader';
import { Context } from '../../hooks/store';
import Modal from '../../components/modal';
import Factory from './factory';

function FactoriesList() {
    const [state, dispatch] = useContext(Context);
    const [factoriesList, setFactoriesList] = useState([]);
    const [modalIsOpen, modalFlip] = useState(false);
    const createFactoryRef = useRef(null);
    const [isLoading, setisLoading] = useState(false);

    const getFactories = async () => {
        setisLoading(true);
        try {
            const data = await httpRequest('GET', ApiEndpoints.GEL_ALL_FACTORIES);
            setFactoriesList(data);
            setisLoading(false);
        } catch (error) {
            setisLoading(false);
        }
    };

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'factories' });
        getFactories();
    }, []);

    useEffect(() => {
        state.socket?.on('factories_overview_data', (data) => {
            setFactoriesList(data);
        });

        setTimeout(() => {
            state.socket?.emit('register_factories_overview_data');
        }, 1000);
        return () => {
            state.socket?.emit('deregister');
        };
    }, [state.socket]);

    const removeFactory = (id) => {
        setFactoriesList(factoriesList.filter((item) => item.id !== id));
    };
    return (
        <div>
            <div className="factories-container">
                <div className="one-edge-shadow">
                    <h1 className="main-header-h1">Factories</h1>
                    <div className="factories-header-flex">
                        <h3>Select a factory to edit</h3>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="36px"
                            placeholder={'Create new factory'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="14px"
                            fontWeight="600"
                            aria-haspopup="true"
                            onClick={() => modalFlip(true)}
                        />
                    </div>
                </div>
                <div className="factories-list">
                    {isLoading && (
                        <div className="loader-uploading">
                            <Loader />
                        </div>
                    )}
                    {factoriesList.map((factory) => {
                        return <Factory key={factory.id} content={factory} removeFactory={() => removeFactory(factory.id)}></Factory>;
                    })}
                    {!isLoading && factoriesList.length === 0 && (
                        <div className="no-factory-to-display">
                            <img src={emptyList} width="100" height="100" alt="emptyList" />
                            <p>There are no factories yet</p>
                            <p className="sub-title">Get started by creating a factory</p>
                            <Button
                                className="modal-btn"
                                width="240px"
                                height="50px"
                                placeholder="Create your first factory"
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
            </div>
            <Modal
                header="Create a factory"
                height="380px"
                width="440px"
                rBtnText="Create"
                lBtnText="Cancel"
                lBtnClick={() => {
                    modalFlip(false);
                }}
                clickOutside={() => modalFlip(false)}
                rBtnClick={() => {
                    createFactoryRef.current();
                    // modalFlip(false);
                }}
                open={modalIsOpen}
            >
                <CreateFactoryDetails createFactoryRef={createFactoryRef} />
            </Modal>
        </div>
    );
}

export default FactoriesList;
